package gotcp

import (
	"context"
	"io"
	net "net"
	"sync"
	"time"

	"github.com/aggronmagi/walle/net/packet"
	"github.com/aggronmagi/walle/net/process"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

// server session
type GoSession struct {
	// process
	*process.Process
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
		err = packet.ErrPacketTooLarge
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
	sess.mux.Lock()
	sess.mux.Unlock()
	if sess.opts.WriteTimeout > 0 {
		sess.conn.SetWriteDeadline(time.Now().Add(sess.opts.WriteTimeout))
	}
	n, err = sess.conn.Write(in)
	if err != nil {
		log.Error3("write message failed", zap.Error(err))
		return
	}

	return
}

func (sess *GoSession) Close() (err error) {
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
	sess.Process.Clean()
}

func (sess *GoSession) writeLoop() {
	log := sess.Process.Opts.Logger
	var err error
	hb := make([][]byte, 0, 16)
	hb = append(hb, sess.opts.PacketHeadBuf())
	defer sess.Close()
	for {
		select {
		case <-sess.ctx.Done():
			for _ = range sess.send {
				// TODO drop message notify
			}
			return
		case data, ok := <-sess.send:
			if !ok {
				sess.conn.Close()
				return
			}
			err = sess.opts.WriteSize(hb[0], len(data))
			if err != nil {
				log.Error3("write message size failed", zap.Error(err))
				return
			}
			buf := net.Buffers{hb[0], data}
			for k := 0; k < len(sess.send); k++ {
				data := <-sess.send
				if k+1 <= len(hb) {
					hb = append(hb, sess.opts.PacketHeadBuf())
				}
				err = sess.opts.WriteSize(hb[k+1], len(data))
				if err != nil {
					log.Error3("write message size failed", zap.Error(err))
					return
				}
				buf = append(buf, hb[k+1])
				buf = append(buf, data)
			}
			if sess.opts.WriteTimeout > 0 {
				sess.conn.SetWriteDeadline(time.Now().Add(sess.opts.WriteTimeout))
			}
			// writev
			_, err = buf.WriteTo(sess.conn)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && (netErr.Timeout() || netErr.Temporary()) {
					// retry
				}
				log.Error3("write data failed", zap.Error(err))
				return
			}
		}
	}
}

func (sess *GoSession) readLoop() {
	log := sess.Process.Opts.Logger
	headSize := len(sess.opts.PacketHeadBuf())
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
				log.Develop8("io.EOF. connect close")
				return
			}
			log.Error3("read head error", zap.Error(err))
			return
		}
		bufSize += read

		for {
			if bufSize < headSize {
				// wait left data
				break
			}
			size := sess.opts.ReadSize(buf[:headSize])
			if bufSize < size+headSize {
				// wait left data
				break
			}
			// NOTE: 复用缓冲区，不能设置Process.DispatchFilter（开启新协程）,或者应该在内部手动复制内存。
			if sess.opts.ReuseReadBuffer {
				sess.Process.OnRead(buf[headSize : headSize+size])
			} else {
				cache := make([]byte, size)
				copy(cache, buf[headSize:headSize+size])
				sess.Process.OnRead(cache)
			}
			copy(buf, buf[headSize+size:])
			bufSize -= headSize + size
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

func (sess *GoSession) AddCloseFunc(f func(sess Session)) {
	sess.closeChain = append(sess.closeChain, f)
}

func (sess *GoSession) newContext(ctx process.Context, ud interface{}) process.Context {
	return &sessionCtx{
		Context:   ctx,
		GoSession: sess,
	}
}

type sessionCtx struct {
	process.Context
	*GoSession
}

var _ SessionContext = &sessionCtx{}
