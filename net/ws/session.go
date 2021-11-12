package ws

import (
	"context"
	"sync"
	"time"

	"github.com/aggronmagi/walle/net/packet"
	"github.com/aggronmagi/walle/net/process"
	"github.com/gorilla/websocket"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

// server session
type WsSession struct {
	// process
	*process.Process
	// websocket conn
	conn *websocket.Conn
	// websocket server
	svr *WsServer
	// session context
	ctx    context.Context
	cancel func()
	// flag
	close atomic.Bool
	send  chan []byte
	mux   sync.Mutex
	//
	writeMethod WriteMethod
	opts        *ServerOptions
	// close call back
	closeChain []func(Session)
}

func (sess *WsSession) Write(in []byte) (n int, err error) {
	// select {
	// case <-sess.ctx.Done():
	// 	err = packet.ErrSessionClosed
	// 	return
	// default:
	// 	// ok
	// }

	if sess.close.Load() {
		err = packet.ErrSessionClosed
		return
	}
	// async write
	if sess.opts.WriteMethods == WriteAsync {
		sess.send <- in
		n = len(in)
		return
	}
	// sync write
	log := sess.Process.Opts.Logger
	n = len(in)
	// websocket write
	sess.mux.Lock()
	defer sess.mux.Unlock()
	if sess.opts.WriteTimeout > 0 {
		sess.conn.SetWriteDeadline(time.Now().Add(sess.opts.WriteTimeout))
	}
	err = sess.conn.WriteMessage(websocket.BinaryMessage, in)
	if err != nil {
		log.Error3("write message failed", zap.Error(err))
		return
	}

	return
}

func (sess *WsSession) Close() (err error) {
	if !sess.close.CAS(false, true) {
		return
	}
	sess.cancel()
	if sess.send != nil {
		close(sess.send)
	}

	return
}

// GetConn get raw conn(net.Conn,websocket.Conn...)
func (sess *WsSession) GetConn() interface{} {
	return sess.conn
}

// GetServer get raw server(*WsServer,*TcpServer...)
func (sess *WsSession) GetServer() Server {
	return sess.svr
}

// Run run client
func (sess *WsSession) Run() {
	// async write
	if sess.opts.WriteMethods == WriteAsync {
		sess.send = make(chan []byte, sess.opts.SendQueueSize)
		go sess.writeLoop()
	}
	sess.readLoop()
	for _, ntf := range sess.closeChain {
		ntf(sess)
	}
	sess.Process.Clean()
}

func (sess *WsSession) writeLoop() {
	defer sess.Close()
	log := sess.Process.Opts.Logger
	tmpChan := make(chan time.Time)
	defer close(tmpChan)
	tickerChan := (<-chan time.Time)(tmpChan)
	if sess.svr != nil && sess.opts.Heartbeat > 0 {
		tickerChan = time.Tick(sess.opts.Heartbeat)
	}

	for {
		select {
		case <-sess.ctx.Done():
			for _ = range sess.send {
				// TODO drop message notify
			}
			return
		case data, ok := <-sess.send:
			if sess.opts.WriteTimeout > 0 {
				sess.conn.SetWriteDeadline(time.Now().Add(sess.opts.WriteTimeout))
			}
			if !ok {
				sess.conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				)
				return
			}
			err := sess.conn.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				log.Error3("write message failed", zap.Error(err))
				return
			}
		case <-tickerChan:
			if sess.opts.WriteTimeout > 0 {
				sess.conn.SetWriteDeadline(time.Now().Add(sess.opts.WriteTimeout))
			}
			if err := sess.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (sess *WsSession) readLoop() {
	defer sess.Close()
	log := sess.Process.Opts.Logger
	if sess.svr != nil {
		if sess.opts.Heartbeat > 0 {
			// FIXME: time set
			sess.conn.SetReadDeadline(time.Now().Add(sess.opts.Heartbeat + time.Second))
			sess.conn.SetPongHandler(func(string) error {
				sess.conn.SetReadDeadline(time.Now().Add(sess.opts.Heartbeat + time.Second))
				return nil
			})
		} else if sess.opts.ReadTimeout > 0 {
			sess.conn.SetReadDeadline(time.Now().Add(sess.opts.ReadTimeout))
		}
	}

	for {
		if sess.svr != nil {
			if sess.opts.Heartbeat == 0 && sess.opts.ReadTimeout > 0 {
				sess.conn.SetReadDeadline(time.Now().Add(sess.opts.ReadTimeout))
			}
		}
		_, data, err := sess.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				log.Error3("read error", zap.Error(err))
			}
			log.Develop8("recv read error", zap.Error(err), zap.Bool("svr", sess.svr != nil))
			return
		}
		err = sess.OnRead(data)
		if err != nil {
			log.Develop8("deal message failed", zap.Error(err))
		}
	}
}

// WithValue wrap context.WithValue
func (sess *WsSession) WithSessionValue(key, value interface{}) {
	sess.ctx = context.WithValue(sess.ctx, key, value)
	return
}

// Value wrap context.Context.Value
func (sess *WsSession) SessionValue(key interface{}) interface{} {
	return sess.ctx.Value(key)
}

func (sess *WsSession) AddCloseFunc(f func(sess Session)) {
	sess.closeChain = append(sess.closeChain, f)
}

func (sess *WsSession) newContext(ctx process.Context, ud interface{}) process.Context {
	return &sessionCtx{
		Context:   ctx,
		WsSession: sess,
	}
}

type sessionCtx struct {
	process.Context
	*WsSession
}

var _ ClientContext = &sessionCtx{}
var _ SessionContext = &sessionCtx{}
