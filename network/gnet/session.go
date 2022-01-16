package gnet

import (
	"context"
	"sync"

	"github.com/aggronmagi/walle/network/rpc"
	process "github.com/aggronmagi/walle/process"
	"github.com/aggronmagi/walle/process/errcode"
	"github.com/panjf2000/gnet/v2"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

// server session
type GNetSession struct {
	// process
	*rpc.RPCProcess
	// websocket conn
	conn gnet.Conn
	// websocket server
	svr *GNetServer
	// session context
	ctx    context.Context
	cancel func()
	// flag
	close atomic.Bool
	udp   bool
	// close call back
	closeChain []func(Session)
}

func (sess *GNetSession) Write(in []byte) (n int, err error) {
	if sess.close.Load() {
		err = errcode.ErrSessionClosed
		return
	}

	// write msg
	n = len(in)
	// if sess.udp {
	// 	err = sess.conn.SendTo(in)
	// } else {
	err = sess.conn.AsyncWrite(in, nil)
	//sess.Opts.FrameLogger.New("gnet.write").Info("session write size", zap.Int("len", n))
	//n, err = sess.conn.Write(in)
	//}
	if err != nil {
		sess.Opts.FrameLogger.New("gnetsesson.Write").Error("write message failed", zap.Error(err))
		return
	}

	return
}

func (sess *GNetSession) Close() (err error) {
	if !sess.close.CAS(false, true) {
		return
	}
	sess.cancel()
	return sess.conn.Close()
}

// GetConn get raw conn(net.Conn,websocket.Conn...)
func (sess *GNetSession) GetConn() interface{} {
	return sess.conn
}

// GetServer get raw server(*WsServer,*TcpServer...)
func (sess *GNetSession) GetServer() Server {
	return sess.svr
}

// WithValue wrap context.WithValue
func (sess *GNetSession) WithSessionValue(key, value interface{}) {
	sess.ctx = context.WithValue(sess.ctx, key, value)
	return
}

// Value wrap context.Context.Value
func (sess *GNetSession) SessionValue(key interface{}) interface{} {
	return sess.ctx.Value(key)
}

func (sess *GNetSession) AddCloseSessionFunc(f func(sess Session)) {
	sess.closeChain = append(sess.closeChain, f)
}

// func (sess *GNetSession) writeLoop() {
// 	sess.send = make(chan []byte, 1024)
// 	for {
// 		select {
// 		case data, ok := <-sess.send:
// 			if !ok {
// 				return
// 			}
// 			sess.conn.AsyncWrite(data)
// 		}
// 	}
// }

func (sess *GNetSession) onClose() {
	for _, ntf := range sess.closeChain {
		ntf(sess)
	}
	sess.RPCProcess.Clean()
}

type sessionCtx struct {
	process.WrapContext
	*GNetSession
}

var _ SessionContext = &sessionCtx{}

// process.ContextPool interface
type gnetServerContextPool struct {
	sync.Pool
}

func (p *gnetServerContextPool) NewContext(inner *process.InnerOptions, opts *process.ProcessOptions, inPkg interface{}, handlers []process.MiddlewareFunc, loadFlag bool) process.Context {
	ctx := p.Get().(*sessionCtx)
	ctx.Inner = inner
	ctx.Opts = opts
	ctx.SrcContext = inner.ParentCtx
	ctx.Index = 0
	ctx.Handlers = handlers
	ctx.InPkg = inPkg
	ctx.LoadFlag = loadFlag
	ctx.Log = opts.Logger
	ctx.FreeContext = ctx
	ctx.GNetSession = inner.BindData.(*GNetSession)
	return ctx
}

func (p *gnetServerContextPool) FreeContext(ctx process.Context) {
	p.Put(ctx)
}

var GNETServerContextPool process.ContextPool = &gnetServerContextPool{
	Pool: sync.Pool{
		New: func() interface{} {
			return &sessionCtx{}
		},
	},
}
