package gotcp

import (
	"context"
	"encoding/binary"
	"io"
	net "net"
	"sync"
	"time"

	"github.com/aggronmagi/walle/network/rpc"
	"github.com/aggronmagi/walle/process"
	"github.com/aggronmagi/walle/process/errcode"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

// server session
type GoSession struct {
	// process
	*rpc.RPCProcess
	// net conn
	conn net.Conn
	// websocket server
	svr *GoServer
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

func (sess *GoSession) Write(in []byte) (n int, err error) {
	if len(in) >= sess.opts.MaxMessageSizeLimit {
		err = errcode.ErrPacketsizeInvalid
		return
	}
	// select {
	// case <-sess.ctx.Done():
	// 	err = packet.ErrSessionClosed
	// 	return
	// default:
	// 	// ok
	// }

	if sess.close.Load() {
		err = errcode.ErrSessionClosed
		return
	}
	// async write
	if sess.opts.WriteMethods == WriteAsync {
		sess.send <- in
		n = len(in)
		return
	}
	// sync write
	sess.mux.Lock()
	sess.mux.Unlock()
	if sess.opts.WriteTimeout > 0 {
		sess.conn.SetWriteDeadline(time.Now().Add(sess.opts.WriteTimeout))
	}
	n, err = sess.conn.Write(in)
	if err != nil {
		sess.opts.FrameLogger.New("goserver.Write").Error("write message failed", zap.Error(err))
		return
	}

	return
}

func (sess *GoSession) Close() (err error) {
	if !sess.close.CAS(false, true) {
		return
	}
	sess.cancel()
	// if sess.send != nil {
	// 	close(sess.send)
	// }

	return
}

// GetConn get raw conn(net.Conn,websocket.Conn...)
func (sess *GoSession) GetConn() interface{} {
	return sess.conn
}

// GetServer get raw server(*WsServer,*TcpServer...)
func (sess *GoSession) GetServer() Server {
	return sess.svr
}

// Run run client
func (sess *GoSession) Run() {
	// async write
	if sess.opts.WriteMethods == WriteAsync {
		sess.send = make(chan []byte, sess.opts.SendQueueSize)
		go sess.writeLoop()
	}
	sess.readLoop()
	for _, ntf := range sess.closeChain {
		ntf(sess)
	}
	sess.Clean()
}

func (sess *GoSession) writeLoop() {
	log := sess.opts.FrameLogger.New("goserver.writeLoop")
	var err error
	var buf net.Buffers = make([][]byte, 0, 32*2)

	defer sess.Close()
	for {
		select {
		case <-sess.ctx.Done():
			for range sess.send {
				// TODO drop message notify
			}
			return
		case data, ok := <-sess.send:
			if !ok {
				sess.conn.Close()
				return
			}

			buf = buf[:0]
			buf = append(buf, data)
			for k := 0; k < len(sess.send); k++ {
				data := <-sess.send
				buf = append(buf, data)
			}
			if sess.opts.WriteTimeout > 0 {
				sess.conn.SetWriteDeadline(time.Now().Add(sess.opts.WriteTimeout))
			}
			// writev
			_, err = buf.WriteTo(sess.conn)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && (netErr.Timeout() || netErr.Temporary()) {
					// TODO: retry
				}
				log.Error("write data failed", zap.Error(err))
				return
			}
		}
	}
}

func (sess *GoSession) readLoop() {
	log := sess.opts.FrameLogger.New("goserver.readLoop")
	buf := make([]byte, sess.opts.ReadBufferSize)
	bufSize := 0
	defer sess.Close()
	for {
		if sess.opts.ReadTimeout > 0 {
			sess.conn.SetReadDeadline(time.Now().Add(sess.opts.ReadTimeout))
		}
		read, err := sess.conn.Read(buf[bufSize:])
		if err != nil {
			if netErr, ok := err.(net.Error); ok && (netErr.Timeout() || netErr.Temporary()) {
				// heatbeat
				continue
			}
			if err == io.EOF {
				log.Warn("io.EOF. connect close")
				return
			}
			log.Error("read head error", zap.Error(err))
			return
		}
		bufSize += read

		for {
			if bufSize < 4 {
				// wait left data
				break
			}
			size := int(binary.BigEndian.Uint32(buf[:4]))
			if bufSize < size+4 {
				// wait left data
				break
			}
			// NOTE: 复用缓冲区，不能设置Process.DispatchFilter（开启新协程）,或者应该在内部手动复制内存。
			if sess.opts.ReuseReadBuffer {
				sess.Process.OnRead(buf[:4+size])
			} else {
				cache := make([]byte, size+4)
				copy(cache, buf[:4+size])
				sess.Process.OnRead(cache)
			}
			copy(buf, buf[4+size:])
			bufSize -= 4 + size
		}
	}
}

// WithValue wrap context.WithValue
func (sess *GoSession) WithSessionValue(key, value interface{}) {
	sess.ctx = context.WithValue(sess.ctx, key, value)
	return
}

// Value wrap context.Context.Value
func (sess *GoSession) SessionValue(key interface{}) interface{} {
	return sess.ctx.Value(key)
}

func (sess *GoSession) AddCloseSessionFunc(f func(sess Session)) {
	sess.closeChain = append(sess.closeChain, f)
}

type sessionCtx struct {
	process.WrapContext
	*GoSession
}

var _ SessionContext = &sessionCtx{}

// process.ContextPool interface
type goServerContextPool struct {
	sync.Pool
}

func (p *goServerContextPool) NewContext(inner *process.InnerOptions, opts *process.ProcessOptions, inPkg interface{}, handlers []process.MiddlewareFunc, loadFlag bool) process.Context {
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
	ctx.GoSession = inner.BindData.(*GoSession)
	return ctx
}

func (p *goServerContextPool) FreeContext(ctx process.Context) {
	p.Put(ctx)
}

var GoServerContextPool process.ContextPool = &goServerContextPool{
	Pool: sync.Pool{
		New: func() interface{} {
			return &sessionCtx{}
		},
	},
}
