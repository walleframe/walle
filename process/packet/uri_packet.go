package packet

import (
	"github.com/walleframe/walle/process/metadata"
	"go.uber.org/zap/zapcore"
)

// Packet define a network packet struct
type Packet struct {
	cmd       PacketCmd  // Packet Cmd
	flag      PacketFlag // Internal flag
	reservd   byte       // Custom Flag
	msgLen    uint8      // MsgUri length
	msgID     uint32
	sessionID uint64 // session id
	msgURI    string
	payload   []byte
	metadata  metadata.MD
	cache     []byte
}

func NewPacket() *Packet {
	return &Packet{
		metadata: make(metadata.MD),
	}
}

func (p *Packet) SetCmd(cmd PacketCmd) {
	p.cmd = cmd
}

func (p *Packet) Cmd() PacketCmd {
	return p.cmd
}

// i = 1-8
func (p *Packet) SetCustomFlag(i byte, set bool) {
	if i < 1 || i > 8 {
		return
	}
	i--
	if set {
		p.reservd |= (1 << i)
	} else {
		p.reservd &= ^(1 << i)
	}
}

func (p *Packet) CustomFlag(i byte) bool {
	if i < 1 || i > 8 {
		return false
	}
	i--
	return (p.reservd & (1 << i)) != 0
}

func (p *Packet) SetFlag(f PacketFlag, set bool) {
	if set {
		p.flag |= f
	} else {
		p.flag &= ^(f)
	}
}

func (p *Packet) HasFlag(f PacketFlag) bool {
	return (p.flag & f) != 0
}

func (p *Packet) SetSeesonID(id uint64) {
	p.sessionID = id
}

func (p *Packet) SessionID() uint64 {
	return p.sessionID
}

func (p *Packet) SetURI(uri string) {
	p.msgURI = uri
}

func (p *Packet) URI() string {
	return p.msgURI
}

func (p *Packet) SetMsgID(id uint32) {
	p.msgID = id
}

func (p *Packet) MsgID() uint32 {
	return p.msgID
}

func (p *Packet) SetMD(md metadata.MD) {
	p.metadata = md
}

func (p *Packet) GetMD() metadata.MD {
	return p.metadata
}

func (p *Packet) Payload() []byte {
	return p.payload
}

func (p *Packet) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	enc.AddInt8("cmd", int8(p.cmd))
	enc.AddUint8("flag", uint8(p.flag))
	enc.AddUint8("reservd", p.reservd)
	enc.AddUint32("mid", p.msgID)
	enc.AddString("uri", p.msgURI)
	enc.AddUint64("sid", p.sessionID)
	enc.AddInt("size", len(p.payload))
	enc.AddObject("md", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) (err error) {
		for k, item := range p.metadata {
			enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
				for _, v := range item {
					enc.AppendString(v)
				}
				return nil
			}))
		}
		return
	}))
	return
}

func NewTestPacket(cmd PacketCmd, payload []byte, md metadata.MD) *Packet {
	return &Packet{
		cmd:      cmd,
		payload:  payload,
		metadata: md,
	}
}

func (p *Packet) CleanForTest() {
	p.cache = nil
	p.msgLen = 0
}
