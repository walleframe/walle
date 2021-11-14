package process

import (
	"context"
	"fmt"
	"sync"

	"github.com/aggronmagi/walle/net/packet"
	"go.uber.org/zap"
)

// rpcSession represents an active calling session.
type rpcSession struct {
	seq uint64
	// sync call use
	done chan *packet.Packet
	// async call use
	aFunc   []RouterFunc
	aReq    *packet.Packet
	aFilter func(ctx Context, req, rsp *packet.Packet)
}

// Process 通用process 封装
type Process struct {
	// rpc session.
	// TODO:待优化, 使用多个锁
	mux        sync.Mutex
	sessionMap map[uint64]*rpcSession
	// config
	Inner *InnerOptions
	Opts  *ProcessOptions
}

func NewProcess(inner *InnerOptions, opts *ProcessOptions) *Process {
	p := &Process{}
	p.Inner = inner
	p.Opts = opts
	return p
}

// OnRead 入口函数。接收数据处理
func (p *Process) OnRead(data []byte) (err error) {
	// dispatch chain
	err = p.Opts.DispatchDataFilter(data, p.innerDealPacket)
	if err != nil {
		p.Opts.Logger.Develop8("dispatch msg failed", zap.Error(err))
	}

	return
}

func (p *Process) innerDealPacket(data []byte) (err error) {
	// 解码网络包
	data = p.Opts.PacketEncode.Decode(data)
	// 反序列化网络包
	pkg := p.Opts.PacketPool.Pop()
	err = p.Opts.PacketCodec.Unmarshal(data, pkg)
	if err != nil {
		p.Opts.Logger.Notice6("unmarshal packet.Paket failed", zap.Error(err))
		return err
	}

	// rpc 请求回包
	if pkg.Cmd == int32(packet.Command_Response) {
		// get and delete session
		sess := p.getDelSession(pkg.Sequence)
		if sess == nil {
			// rpc 已超时
			p.Opts.Logger.Develop8("rpc respond session not found", zap.Object("pkg", pkg))
			return
		}
		// Async Call
		if len(sess.aFunc) > 0 {
			wrapCtx := &wrapContext{}
			wrapCtx.p = p
			wrapCtx.log = p.Opts.Logger
			wrapCtx.src = p.Inner.ParentCtx
			wrapCtx.in = pkg
			wrapCtx.handlers = sess.aFunc
			// aFilter -> ctx.Next(ctx)
			// NOTE: sess.aReq only valid in aFilter.
			sess.aFilter(p.Inner.NewContext(wrapCtx, p.Inner.BindData), sess.aReq, pkg)
			// free req packet.
			p.Opts.PacketPool.Push(sess.aReq)
			return
		}
		// Sync Call
		sess.done <- pkg
		return
	}
	if p.Inner.Router == nil {
		err = packet.ErrUnexpectedCode
		p.Opts.Logger.Notice6("unexcepted code: not set Router)", zap.Object("pkg", pkg))
		p.Opts.PacketPool.Push(pkg)
		return
	}

	// Request or Notice
	handlers, err := p.Inner.Router.GetHandlers(pkg)
	if err != nil {
		p.Opts.Logger.Notice6("get handler failed", zap.Object("pkg", pkg), zap.Error(err))
		p.Opts.PacketPool.Push(pkg)
		return err
	}

	wrapCtx := &wrapContext{}
	wrapCtx.p = p
	wrapCtx.log = p.Opts.Logger
	wrapCtx.src = p.Inner.ParentCtx
	wrapCtx.in = pkg
	wrapCtx.handlers = handlers

	ctx := p.Inner.NewContext(wrapCtx, p.Inner.BindData)
	// load limit
	if p.Opts.LoadLimitFilter(ctx, p.Inner.Load.Add(1), wrapCtx.in) {
		p.Opts.PacketPool.Push(pkg)
		p.Inner.Load.Dec()
		p.Opts.Logger.Develop8("process load limit", zap.Object("pkg", pkg))
		return
	}
	// Note: p.load.Decr() by Context
	wrapCtx.loadFlag = true
	ctx.Next(ctx)

	return
}

func (p *Process) Call(ctx context.Context, uri interface{}, rq, rs interface{}, opts *CallOptions) (err error) {
	if p.Inner.Output == nil {
		err = packet.ErrUnexpectedCode
		p.Opts.Logger.Develop8("unexcepted code: not set Output(io.Writer)", zap.Any("uri", uri))
		return
	}
	req, err := p.NewPacket(packet.Command_Request, uri, rq, opts.Metadata)
	if err != nil {
		return
	}
	defer p.Opts.PacketPool.Push(req)

	session := &rpcSession{
		seq:  req.Sequence,
		done: make(chan *packet.Packet, 1),
	}

	p.saveSession(req.Sequence, session)
	defer p.delSession(req.Sequence)

	// timeout options
	if opts.Timeout > 0 {
		nctx, cancel := context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
		ctx = nctx
	}

	err = p.WritePacket(ctx, req)
	if err != nil {
		p.Opts.Logger.Develop8("write data failed", zap.Error(err), zap.Any("reqeust", rq), zap.Object("packet", req))
		return
	}

	select {
	case <-ctx.Done():
		err = packet.ErrTimeout
		p.Opts.Logger.Develop8("request rpc timeout", zap.Error(err), zap.Any("reqeust", rq), zap.Object("packet", req))
		return
	case rsp, ok := <-session.done:
		if !ok {
			p.Opts.Logger.Develop8("session closed", zap.Error(err), zap.Any("reqeust", rq), zap.Object("packet", req))
			err = packet.ErrTimeout
			break
		}
		defer p.Opts.PacketPool.Push(rsp)
		// unmarshal call response
		err = p.unmarshalRspBody(rsp, rs)
	}

	return
}

func (p *Process) AsyncCall(ctx context.Context, uri interface{}, rq interface{}, af RouterFunc, opts *AsyncCallOptions) (err error) {
	if p.Inner.Output == nil {
		err = packet.ErrUnexpectedCode
		p.Opts.Logger.Develop8("unexcepted code: not set Output(io.Writer)", zap.Any("uri", uri))
		return
	}
	req, err := p.NewPacket(packet.Command_Request, uri, rq, opts.Metadata)
	if err != nil {
		return
	}

	req.Flag |= uint32(packet.Flag_ClientAsync)

	session := &rpcSession{
		seq: req.Sequence,
	}
	// FIXME: use middleware??? {
	// session.async = make([]RouterFunc, len(p.router.middlewares)+1)
	// copy(session.async, p.router.middlewares)
	// session.async[len(session.async)-1] = af
	// }
	session.aFunc = append(session.aFunc, af)
	session.aFilter = opts.ResponseFilter
	session.aReq = req
	p.saveSession(req.Sequence, session)

	// timeout options
	var cancel func()
	if opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
	}
	err = p.WritePacket(ctx, req)
	if err != nil {
		p.delSession(req.Sequence)
		p.Opts.PacketPool.Push(req)
		p.Opts.Logger.Develop8("write data failed", zap.Error(err), zap.Any("reqeust", rq), zap.Object("packet", req))
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
			if last := p.getDelSession(req.Sequence); last != nil {
				p.Opts.Logger.Develop8("async request rpc timeout", zap.Any("reqeust", rq), zap.Object("packet", req))
				wrapCtx := &wrapContext{}
				wrapCtx.p = p
				wrapCtx.log = p.Opts.Logger
				wrapCtx.src = p.Inner.ParentCtx
				wrapCtx.in, _ = p.NewPacket(packet.Command_Response, uri, packet.ErrTimeout, nil, true)
				wrapCtx.handlers = last.aFunc
				// aFilter -> ctx.Next(ctx)
				// NOTE: sess.aReq only valid in aFilter.
				last.aFilter(p.Inner.NewContext(wrapCtx, p.Inner.BindData), last.aReq, wrapCtx.in)
				p.Opts.PacketPool.Push(req)
			}
		}
	})

	return
}

func (p *Process) Notify(ctx context.Context, uri interface{}, rq interface{}, opts *NoticeOptions) (err error) {
	if p.Inner.Output == nil {
		err = packet.ErrUnexpectedCode
		p.Opts.Logger.Develop8("unexcepted code: not set Output(io.Writer)", zap.Any("uri", uri))
		return
	}
	req, err := p.NewPacket(packet.Command_Oneway, uri, rq, opts.Metadata)
	if err != nil {
		return
	}
	defer p.Opts.PacketPool.Push(req)

	// timeout options
	if opts.Timeout > 0 {
		nctx, cancel := context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
		ctx = nctx
	}
	err = p.WritePacket(ctx, req)
	if err != nil {
		p.Opts.Logger.Develop8("write data failed", zap.Error(err), zap.Any("reqeust", rq), zap.Object("packet", req))
		return
	}

	return
}

func (p *Process) Clean() {
	data, _ := p.Opts.MsgCodec.Marshal(packet.ErrSessionClosed)
	p.mux.Lock()
	defer p.mux.Unlock()
	if p.sessionMap == nil {
		return
	}
	fmt.Println("clean process")
	// clean rpc session
	for _, sess := range p.sessionMap {
		// Async Call
		if len(sess.aFunc) > 0 {
			pkg := p.Opts.PacketPool.Pop()
			*pkg = *sess.aReq
			pkg.Flag = uint32(packet.Flag_Exception)
			pkg.Body = data
			wrapCtx := &wrapContext{}
			wrapCtx.p = p
			wrapCtx.log = p.Opts.Logger
			wrapCtx.src = p.Inner.ParentCtx
			wrapCtx.in = pkg
			wrapCtx.handlers = sess.aFunc
			// aFilter -> ctx.Next(ctx)
			// NOTE: sess.aReq only valid in aFilter.
			sess.aFilter(p.Inner.NewContext(wrapCtx, p.Inner.BindData), sess.aReq, pkg)
			// free req packet.
			p.Opts.PacketPool.Push(sess.aReq)
			continue
		}
		// Sync Call
		close(sess.done)
	}
	p.sessionMap = nil
}

// func (p *Process) WriteMessage(ctx Context, msg interface{}, md ...MetadataOption) (err error) {
// 	if p.Inner.Output == nil {
// 		err = packet.ErrUnexpectedCode
// 		p.Opts.Logger.Error3("unexcepted code: not set Output(io.Writer)", zap.Any("msg", msg))
// 		return
// 	}
// 	req, err := p.newPacket(packet.Command_Oneway, md, msg)
// 	if err != nil {
// 		return
// 	}
// 	defer p.Opts.PacketPool.Push(req)
// 	return p.WritePacket(ctx, req)
// }

func (p *Process) WritePacket(ctx context.Context, req *packet.Packet) (err error) {
	if p.Inner.Output == nil {
		err = packet.ErrUnexpectedCode
		p.Opts.Logger.Error3("unexcepted code: not set Output(io.Writer)", zap.Object("packet", req))
		return
	}
	data, err := p.MarshalPacket(req)
	if err != nil {
		p.Opts.Logger.Develop8("marshal packet failed", zap.Error(err), zap.Object("packet", req))
		return
	}

	// opts.Out 处理连接状态.
	_, err = p.Inner.Output.Write(data)
	if err != nil {
		p.Opts.Logger.Develop8("io.write failed", zap.Error(err), zap.Object("packet", req))
		return
	}
	return
}

func (p *Process) MarshalPacket(req *packet.Packet) (data []byte, err error) {
	data, err = p.Opts.PacketCodec.Marshal(req)
	if err != nil {
		p.Opts.Logger.Develop8("marshal message failed", zap.Error(err), zap.Object("packet", req))
		return
	}

	if req.Sequence == 0 {
		req.Sequence = uint64(p.Inner.Sequence.Add(1))
	}

	data = p.Opts.PacketEncode.Encode(data)
	return
}

// func (p *Process) Write(data []byte) (n int, err error) {
// 	if p.Inner.Output == nil {
// 		err = packet.ErrUnexpectedCode
// 		p.Opts.Logger.Develop8("unexcepted code: not set Output(io.Writer)")
// 		return
// 	}
// 	return p.Inner.Output.Write(data)
// }

func (p *Process) NewPacket(cmd packet.Command, uri, rq interface{}, md []MetadataOption, errflag ...bool) (req *packet.Packet, err error) {
	// TODO: 优化. 预申请buffer
	req = p.Opts.PacketPool.Pop()
	req.Cmd = int32(cmd)
	if len(md) > 0 {
		req.Metadata = make(map[string]string, len(md))
		for _, v := range md {
			v(req)
		}
	}
	if len(errflag) > 0 && errflag[0] {
		req.Flag |= uint32(packet.Flag_Exception)
	}
	req.Sequence = uint64(p.Inner.Sequence.Add(1))
	switch v := uri.(type) {
	case uint32:
		req.ReservedRq = v
	case string:
		req.Uri = v
	default:
		err = packet.ErrUnexpectedCode
		p.Opts.Logger.Develop8("unexcepted code: uri invaliid type.",
			zap.Any("reqeust", rq), zap.Any("uri", uri),
		)
		return
	}
	req.Body, err = p.Opts.MsgCodec.Marshal(rq)
	if err != nil {
		p.Opts.Logger.Develop8("marshal message failed", zap.Error(err), zap.Any("reqeust", rq))
		p.Opts.PacketPool.Push(req)
		return
	}
	return
}

func (p *Process) NewResponse(in *packet.Packet, body interface{}, md []MetadataOption) (rsp *packet.Packet, err error) {
	if in.Cmd != int32(packet.Command_Request) {
		err = packet.ErrUnexpectedCode
		p.Opts.Logger.Develop8("unexcepted code: not request packet.",
			zap.Object("in", in),
		)
		return
	}

	// TODO: 优化. 预申请buffer
	rsp = p.Opts.PacketPool.Pop()
	rsp.Cmd = int32(packet.Command_Response)
	rsp.Flag = in.Flag
	rsp.Sequence = in.Sequence
	// rsp.Metadata = p.Metadata
	rsp.ReservedRq = in.ReservedRq
	rsp.Uri = in.Uri
	if len(md) > 0 {
		rsp.Metadata = make(map[string]string, len(md))
		for _, v := range md {
			v(rsp)
		}
	}

	var rb interface{}

	switch v := body.(type) {
	case *packet.ErrorResponse:
		rb = v
		rsp.SetErrorFlag(true)
	case error:
		rb = packet.ErrUnkown.WrapError(v)
		rsp.SetErrorFlag(true)
	default:
		rb = v
	}
	rsp.Body, err = p.Opts.MsgCodec.Marshal(rb)
	if err != nil {
		p.Opts.Logger.Develop8("marshal respond message failed", zap.Error(err))
		p.Opts.PacketPool.Push(rsp)
		return
	}
	return
}

func (p *Process) getDelSession(id uint64) (sess *rpcSession) {
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

func (p *Process) delSession(id uint64) (sess *rpcSession) {
	p.mux.Lock()
	if p.sessionMap == nil {
		p.sessionMap = make(map[uint64]*rpcSession)
	}
	delete(p.sessionMap, id)
	p.mux.Unlock()
	return
}

func (p *Process) saveSession(id uint64, sess *rpcSession) {
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

func (p *Process) unmarshalRspBody(pack *packet.Packet, rs interface{}) error {
	if len(pack.Body) < 1 {
		return nil
	}
	// not error response
	if !pack.HasFlag(packet.Flag_Exception) {
		if rs == nil {
			return nil
		}
		return p.Opts.MsgCodec.Unmarshal(pack.Body, rs)
	}
	// FIXME: use pool ? where to free ?
	errMsg := new(packet.ErrorResponse)
	err2 := p.Opts.MsgCodec.Unmarshal(pack.Body, errMsg)
	if err2 != nil {
		return err2
	}
	return errMsg
}
