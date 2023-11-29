package rpc

import (
	"context"
	"sync"

	"github.com/walleframe/walle/process"
	"github.com/walleframe/walle/process/errcode"
	"github.com/walleframe/walle/process/packet"
	"github.com/walleframe/walle/zaplog"
	"go.uber.org/zap"
)

//go:generate gogen imake . -t=RPCProcess -r RPCProcess=RPCProcesser  -o rpc_processer.go --merge
//go:generate mockgen -source rpc_processer.go -destination ../../testpkg/mock_rpc/processer.go

// rpcSession represents an active calling session.
type rpcSession struct {
	seq uint64
	// sync call use
	done chan *packet.Packet
	// async call use
	aFunc   []process.RouterFunc
	aReq    *packet.Packet
	aFilter func(ctx process.Context, req, rsp interface{})
}

// RPCProcess 通用rpc处理流程封装 封装
type RPCProcess struct {
	process.Process
	// rpc session.
	// TODO:待优化, 使用多个锁
	mux        sync.Mutex
	sessionMap map[uint64]*rpcSession
}

func NewRPCProcess(inner *process.InnerOptions, opts *process.ProcessOptions) *RPCProcess {
	p := &RPCProcess{
		Process: process.NewProcess(inner, opts),
	}
	p.Process.Filter = p.OnReply
	return p
}

func (p *RPCProcess) logger(fname string) *zaplog.LogEntities {
	return p.Opts.FrameLogger.New(fname)
}

// OnReply rpc请求返回处理
func (p *RPCProcess) OnReply(in interface{}) (filter bool) {
	rsp, ok := in.(*packet.Packet)
	if !ok {
		return
	}
	// rpc 请求回包
	if rsp.Cmd() != packet.CmdResponse {
		return
	}
	filter = true
	log := p.logger("rpcprocess.OnReply")
	// get and delete session
	sess := p.getDelSession(rsp.SessionID())
	if sess == nil {
		// rpc 已超时
		log.Error("rpc respond session not found", zap.Object("pkg", rsp))
		return
	}
	// Async Call without timeout deal
	if sess.done == nil && len(sess.aFunc) > 0 {
		// aFilter -> ctx.Next(ctx)
		// NOTE: sess.aReq only valid in aFilter.
		sess.aFilter(p.Inner.ContextPool.NewContext(p.Inner, p.Opts, rsp, sess.aFunc, false), sess.aReq, rsp)
		// free req packet.
		p.Opts.PacketPool.Put(sess.aReq)
		return
	}
	// Sync Call or Async Call with timeout.
	sess.done <- rsp
	return
}

// Call 同步rpc请求
func (p *RPCProcess) Call(ctx context.Context, uri interface{}, rq, rs interface{}, opts *CallOptions) (err error) {
	log := p.logger("process.Call")
	if p.Inner.Output == nil {
		err = errcode.ErrUnexpectedCode
		log.Debug("unexcepted code: not set Output(io.Writer)", zap.Any("uri", uri))
		return
	}

	req := p.Opts.PacketPool.Get().(*packet.Packet)
	defer p.Opts.PacketPool.Put(req)
	err = p.Opts.PacketWraper.NewPacket(req, packet.CmdRequest, uri, opts.Metadata)
	if err != nil {
		log.Error("new packet failed", zap.Error(err), zap.Any("reqeust", rq), zap.Object("packet", req))
		return
	}
	err = p.Opts.PacketWraper.PayloadMarshal(req, p.Opts.MsgCodec, rq)
	if err != nil {
		log.Error("marshal payload failed", zap.Error(err), zap.Any("reqeust", rq), zap.Object("packet", req))
		return
	}

	data, err := p.Opts.PacketCodec.Marshal(req)
	if err != nil {
		log.Error("marshal packet failed", zap.Error(err), zap.Any("reqeust", rq), zap.Object("packet", req))
		return
	}

	data = p.Opts.PacketEncode.Decode(data)

	session := &rpcSession{
		seq:  req.SessionID(),
		done: make(chan *packet.Packet, 1),
	}

	p.saveSession(req.SessionID(), session)
	defer p.delSession(req.SessionID())

	// timeout options
	if opts.Timeout > 0 {
		nctx, cancel := context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
		ctx = nctx
	}

	// send request
	_, err = p.Inner.Output.Write(data)
	if err != nil {
		log.Error("write data failed", zap.Error(err), zap.Any("reqeust", rq), zap.Object("packet", req))
		return
	}

	select {
	case <-ctx.Done():
		err = errcode.ErrTimeout
		log.Warn("request rpc timeout", zap.Error(err), zap.Any("reqeust", rq), zap.Object("packet", req))
		return
	case rsp, ok := <-session.done:
		if !ok {
			log.Warn("session closed", zap.Error(err), zap.Any("reqeust", rq), zap.Object("packet", req))
			err = errcode.ErrTimeout
			break
		}
		defer p.Opts.PacketPool.Put(rsp)
		// unmarshal call response
		err = p.Opts.PacketWraper.PayloadUnmarshal(rsp, p.Opts.MsgCodec, rs)
	}

	return
}

// AsyncCall 异步RPC请求
func (p *RPCProcess) AsyncCall(ctx context.Context, uri interface{}, rq interface{}, af process.RouterFunc, opts *AsyncCallOptions) (err error) {
	log := p.logger("process.AsyncCall")
	if p.Inner.Output == nil {
		err = errcode.ErrUnexpectedCode
		log.Error("unexcepted code: not set Output(io.Writer)", zap.Any("uri", uri))
		return
	}
	req := p.Opts.PacketPool.Get().(*packet.Packet)
	err = p.Opts.PacketWraper.NewPacket(req, packet.CmdRequest, uri, opts.Metadata)
	if err != nil {
		log.Error("new packet failed", zap.Error(err), zap.Any("reqeust", rq), zap.Object("packet", req))
		return
	}
	err = p.Opts.PacketWraper.PayloadMarshal(req, p.Opts.MsgCodec, rq)
	if err != nil {
		log.Error("marshal payload failed", zap.Error(err), zap.Any("reqeust", rq), zap.Object("packet", req))
		return
	}

	data, err := p.Opts.PacketCodec.Marshal(req)
	if err != nil {
		log.Error("marshal packet failed", zap.Error(err), zap.Any("reqeust", rq), zap.Object("packet", req))
		return
	}

	data = p.Opts.PacketEncode.Decode(data)

	session := &rpcSession{
		seq: req.SessionID(),
	}
	// FIXME: use middleware??? {
	// session.async = make([]RouterFunc, len(p.router.middlewares)+1)
	// copy(session.async, p.router.middlewares)
	// session.async[len(session.async)-1] = af
	// }
	session.aFunc = append(session.aFunc, af)
	session.aFilter = opts.ResponseFilter
	session.aReq = req
	p.saveSession(req.SessionID(), session)
	sessionID := req.SessionID()

	// timeout options
	var cancel func()
	if opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		// with timeout need use response chan
		session.done = make(chan *packet.Packet, 1)
	}
	// send request
	_, err = p.Inner.Output.Write(data)
	if err != nil {
		cancel()
		p.delSession(req.SessionID())
		p.Opts.PacketPool.Put(req)
		log.Error("write data failed", zap.Error(err), zap.Any("reqeust", rq), zap.Object("packet", req))
		return
	}

	if cancel == nil {
		return
	}
	// async wait. default run wait in another gorountine
	opts.WaitFilter(func() {
		defer cancel()
		select {
		case <-ctx.Done():
			p.asyncCallTimeout(sessionID)
		case rsp := <-session.done:
			// aFilter -> ctx.Next(ctx)
			// NOTE: sess.aReq only valid in aFilter.
			session.aFilter(p.Inner.ContextPool.NewContext(p.Inner, p.Opts, rsp, session.aFunc, false), session.aReq, rsp)
			// free req packet.
			p.Opts.PacketPool.Put(session.aReq)
		}
	})

	return
}

func (p *RPCProcess) asyncCallTimeout(sessionId uint64) {
	log := p.logger("rpcprocess.asyncCallTimeout")
	var err error
	last := p.getDelSession(sessionId)
	if last == nil {
		log.Debug("allready deal request", zap.Uint64("sequenceID", sessionId))
		return
	}
	defer p.Opts.PacketPool.Put(last.aReq)
	log.Warn("async request rpc timeout", zap.Object("packet", last.aReq))
	rsp := p.Opts.PacketPool.Get().(*packet.Packet)
	err = p.Opts.PacketWraper.NewResponse(last.aReq, rsp, nil)
	if err != nil {
		log.Error("new timeout packet failed", zap.Error(err), zap.Object("packet", last.aReq))
		return
	}
	err = p.Opts.PacketWraper.PayloadMarshal(rsp, p.Opts.MsgCodec, errcode.ErrTimeout)
	if err != nil {
		log.Error("marshal timeout payload failed", zap.Error(err), zap.Object("packet", last.aReq))
		return
	}

	// aFilter -> ctx.Next(ctx)
	// NOTE: sess.aReq only valid in aFilter.
	last.aFilter(p.Inner.ContextPool.NewContext(p.Inner, p.Opts, rsp, last.aFunc, false), last.aReq, rsp)
}

// Notify 通知请求(one way)
func (p *RPCProcess) Notify(ctx context.Context, uri interface{}, rq interface{}, opts *NoticeOptions) (err error) {
	log := p.logger("process.Notify")
	if p.Inner.Output == nil {
		err = errcode.ErrUnexpectedCode
		log.Error("unexcepted code: not set Output(io.Writer)", zap.Any("uri", uri))
		return
	}
	req := p.Opts.PacketPool.Get().(*packet.Packet)
	err = p.Opts.PacketWraper.NewPacket(req, packet.CmdRequest, uri, opts.Metadata)
	if err != nil {
		log.Error("new packet failed", zap.Error(err), zap.Any("reqeust", rq), zap.Object("packet", req))
		return
	}
	err = p.Opts.PacketWraper.PayloadMarshal(req, p.Opts.MsgCodec, rq)
	if err != nil {
		log.Error("marshal payload failed", zap.Error(err), zap.Any("reqeust", rq), zap.Object("packet", req))
		return
	}

	data, err := p.Opts.PacketCodec.Marshal(req)
	if err != nil {
		log.Error("marshal packet failed", zap.Error(err), zap.Any("reqeust", rq), zap.Object("packet", req))
		return
	}

	data = p.Opts.PacketEncode.Decode(data)

	// timeout options
	if opts.Timeout > 0 {
		nctx, cancel := context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
		ctx = nctx
	}
	// send request
	_, err = p.Inner.Output.Write(data)
	if err != nil {
		log.Error("write data failed", zap.Error(err), zap.Any("reqeust", rq), zap.Object("packet", req))
		return
	}

	return
}

// Clean session 清理（rpc请求等缓存清理）
func (p *RPCProcess) Clean() {
	p.mux.Lock()
	defer p.mux.Unlock()
	if p.sessionMap == nil {
		return
	}
	// clean rpc session
	for _, sess := range p.sessionMap {
		// Async Call
		if len(sess.aFunc) > 0 {
			rsp := p.Opts.PacketPool.Get()
			p.Opts.PacketWraper.NewResponse(sess.aReq, rsp, nil)
			//packet.Command_Response, sess.aReq.Uri, packet.ErrTimeout, nil, true
			p.Opts.PacketWraper.PayloadMarshal(rsp, p.Opts.MsgCodec, errcode.ErrTimeout)

			// aFilter -> ctx.Next(ctx)
			// NOTE: sess.aReq only valid in aFilter.
			sess.aFilter(p.Inner.ContextPool.NewContext(p.Inner, p.Opts, rsp, sess.aFunc, false), sess.aReq, rsp)
			// free req packet.
			p.Opts.PacketPool.Put(sess.aReq)
			continue
		}
		// Sync Call
		close(sess.done)
	}
	p.sessionMap = nil
}

func (p *RPCProcess) getDelSession(id uint64) (sess *rpcSession) {
	p.mux.Lock()
	if p.sessionMap == nil {
		p.sessionMap = make(map[uint64]*rpcSession)
	}
	if last, ok := p.sessionMap[id]; ok {
		sess = last
		delete(p.sessionMap, id)
	}
	p.mux.Unlock()
	return
}

func (p *RPCProcess) delSession(id uint64) (sess *rpcSession) {
	p.mux.Lock()
	if p.sessionMap == nil {
		p.sessionMap = make(map[uint64]*rpcSession)
	}
	delete(p.sessionMap, id)
	p.mux.Unlock()
	return
}

func (p *RPCProcess) saveSession(id uint64, sess *rpcSession) {
	p.mux.Lock()
	if p.sessionMap == nil {
		p.sessionMap = make(map[uint64]*rpcSession)
	}
	if _, ok := p.sessionMap[id]; ok {
		panic("same session id")
	}
	p.sessionMap[id] = sess
	p.mux.Unlock()
	return
}
