package gnet

import (
	"context"

	"github.com/aggronmagi/walle/net/packet"
	process "github.com/aggronmagi/walle/net/process"
	"github.com/panjf2000/gnet"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

// server session
type GNetSession struct {
	// process
	*process.Process
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
		err = packet.ErrSessionClosed
		return
	}

	// write msg
	log := sess.Process.Opts.Logger
	n = len(in)
	if sess.udp {
		err = sess.conn.SendTo(in)
	} else {
		err = sess.conn.AsyncWrite(in)
	}
	if err != nil {
		log.Error3("write message failed", zap.Error(err))
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

func (sess *GNetSession) AddCloseFunc(f func(sess Session)) {
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

func (sess *GNetSession) newContext(ctx process.Context, ud interface{}) process.Context {
	return &sessionCtx{
		Context:     ctx,
		GNetSession: sess,
	}
}

func (sess *GNetSession) onClose() {
	for _, ntf := range sess.closeChain {
		ntf(sess)
	}
	sess.Process.Clean()
}

type sessionCtx struct {
	process.Context
	*GNetSession
}

var _ SessionContext = &sessionCtx{}
