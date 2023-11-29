package packet

import (
	"github.com/walleframe/walle/process/errcode"
	"github.com/walleframe/walle/process/message"
	"github.com/walleframe/walle/process/metadata"
	atomic "go.uber.org/atomic"
)

var defaultPacketWraper ProtocolWraper = &packetProtocolWraper{
	sequence: atomic.Int64{},
}

// GetPacketPool get default packet pool
func GetProtocolWraper() ProtocolWraper {
	return defaultPacketWraper
}

// SetPacketPool set default packet pool
func SetPacketWraper(p ProtocolWraper) {
	defaultPacketWraper = p
}

func NewPacketWraper() ProtocolWraper {
	return &packetProtocolWraper{
		sequence: atomic.Int64{},
	}
}

type packetProtocolWraper struct {
	sequence atomic.Int64
}

// unmarshal packet's payload by msg codec
func (w *packetProtocolWraper) PayloadUnmarshal(pkg interface{}, codec message.Codec, obj interface{}) error {
	p, ok := pkg.(*Packet)
	if !ok {
		return errcode.ErrUnexpectedCode
	}
	if !p.HasFlag(FlagError) {
		// support call reponse is nil
		if obj == nil {
			return nil
		}
		return codec.Unmarshal(p.payload, obj)
	}
	return errcode.DefaultErrorCodec.Unmarshal(p.payload)
}

// marshal packet's payload by msg codec,then set payload binary data into message buf.
func (w *packetProtocolWraper) PayloadMarshal(pkg interface{}, codec message.Codec, payload interface{}) (err error) {
	p, ok := pkg.(*Packet)
	if !ok {
		err = errcode.ErrUnexpectedCode
		return
	}
	// if p.payload != nil {
	// 	mempool.Pool().Free(p.payload)
	// }
	// support call reponse is nil
	if payload == nil {
		return
	}
	switch v := payload.(type) {
	case error:
		p.payload, err = errcode.DefaultErrorCodec.Marshal(v)
		p.SetFlag(FlagError, true)
	default:
		p.payload, err = codec.Marshal(payload)
	}

	return
}

// new response packet
func (w *packetProtocolWraper) NewResponse(inPkg, outPkg interface{}, md metadata.MD) (err error) {
	req, ok := inPkg.(*Packet)
	if !ok {
		err = errcode.ErrUnexpectedCode
		return
	}
	rsp, ok := outPkg.(*Packet)
	if !ok {
		err = errcode.ErrUnexpectedCode
		return
	}
	rsp.SetCmd(CmdResponse)
	rsp.flag = 0
	rsp.reservd = req.reservd
	rsp.sessionID = req.sessionID
	rsp.msgURI = req.msgURI
	rsp.msgID = req.msgID
	rsp.metadata = md
	return
}

func (w *packetProtocolWraper) NewPacket(inPkg interface{}, cmd PacketCmd, uri interface{}, md metadata.MD) (err error) {
	req, ok := inPkg.(*Packet)
	if !ok {
		err = errcode.ErrUnexpectedCode
		return
	}
	req.cmd = cmd
	switch v := uri.(type) {
	case uint32:
		req.msgID = v
	case string:
		req.msgURI = v
	default:
		err = errcode.ErrUnexpectedCode
		return
	}
	req.sessionID = uint64(w.sequence.Add(1))
	if len(md) > 0 {
		req.metadata = md
	}
	return
}

// get pkg metadata
func (w *packetProtocolWraper) GetMetadata(pkg interface{}) (md metadata.MD, err error) {
	req, ok := pkg.(*Packet)
	if !ok {
		err = errcode.ErrUnexpectedCode
		return
	}
	md = req.metadata
	return
}
