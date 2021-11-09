package packet

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"go.uber.org/zap/zapcore"
)

// type DebugLogPacket *Packet

// func (x *DebugLogPacket) MarshalLogObject(e zapcore.ObjectEncoder) (err error) {
// 	e.AddString("cmd", Flag(x.Cmd).String())
// 	e.Add
// 	return
// }

func (x *Packet) MarshalLogObject(e zapcore.ObjectEncoder) (err error) {
	e.AddInt32("cmd", x.Cmd)
	e.AddUint32("flag", x.Flag)
	e.AddUint64("sequence", x.Sequence)
	e.AddObject("metadata", zapcore.ObjectMarshalerFunc(
		func(oe zapcore.ObjectEncoder) error {
			for k, v := range x.Metadata {
				oe.AddString(k, v)
			}
			return nil
		},
	))
	e.AddString("uri", x.Uri)
	e.AddUint32("rqid", x.ReservedRq)
	e.AddString("body", base64.StdEncoding.EncodeToString(x.Body))
	//e.AddUint32("r1", x.ReservedUint32)
	return
}

func (p *Packet) NewResponse() (rsp *Packet) {
	if p.Cmd != int32(Command_Request) {
		return
	}
	rsp = DefaultPacketPool.Pop()
	rsp.Cmd = int32(Command_Response)
	rsp.Flag = p.Flag
	rsp.Sequence = p.Sequence
	// rsp.Metadata = p.Metadata
	rsp.ReservedRq = p.ReservedRq
	rsp.Uri = p.Uri
	// rsp.ReservedInt32 = p.ReservedInt32
	// rsp.ReservedUint64 = p.ReservedUint64
	// rsp.ReservedString = p.ReservedString
	// rsp.ReservedBytes = p.ReservedBytes
	return
}

// MarkFlag set mark flag
func (p *Packet) MarkFlag(flag Flag) {
	if p == nil {
		return
	}
	p.Flag |= uint32(flag)
}

// UnmarkFlag unmark flag
func (p *Packet) UnmarkFlag(flag Flag) {
	if p == nil {
		return
	}
	p.Flag &= ^uint32(flag)
}

// HasFlag identifies whether has marked flag
func (p *Packet) HasFlag(flag Flag) bool {
	if p == nil {
		return false
	}
	return (p.Flag & uint32(flag)) > 0
}

func (p *Packet) SetErrorFlag(set bool) {
	if set {
		p.MarkFlag(Flag_Exception)
	} else {
		p.UnmarkFlag(Flag_Exception)
	}
}

// GetMetadataString get metadata by key
func (p *Packet) GetMetadataString(key string) (val string, ok bool) {
	if p == nil || p.Metadata == nil {
		return
	}
	val, ok = p.Metadata[key]
	return
}

// GetMetadataBytes get metadata by key
func (p *Packet) GetMetadataBytes(key string) (val []byte, ok bool) {
	if p == nil || p.Metadata == nil {
		return
	}
	tmp, ok := p.Metadata[key]
	if !ok {
		return
	}
	val = []byte(tmp)
	return
}

// GetMetadataUint64 get metadata by key and convert to uint64
func (p *Packet) GetMetadataUint64(key string) (val uint64, ok bool) {
	if p == nil || p.Metadata == nil {
		return
	}
	tmp, ok := p.Metadata[key]
	if !ok {
		return
	}
	val, err := strconv.ParseUint(tmp, 10, 64)
	if err != nil {
		ok = false
		val = 0
		return
	}
	return
}

// GetMetadataInt64 get metadata by key and convert to int64
func (p *Packet) GetMetadataInt64(key string) (val int64, ok bool) {
	if p == nil || p.Metadata == nil {
		return
	}
	tmp, ok := p.Metadata[key]
	if !ok {
		return
	}
	val, err := strconv.ParseInt(tmp, 10, 64)
	if err != nil {
		ok = false
		val = 0
		return
	}
	return
}

// SetMetadataString set metadata string value
func (p *Packet) SetMetadataString(key, val string) {
	if p == nil {
		return
	}
	if p.Metadata == nil {
		p.Metadata = make(map[string]string)
	}
	p.Metadata[key] = val
	return
}

// SetMetadataBytes set metadata bytes value
func (p *Packet) SetMetadataBytes(key string, val []byte) {
	if p == nil {
		return
	}
	if p.Metadata == nil {
		p.Metadata = make(map[string]string)
	}
	p.Metadata[key] = string(val)
	return
}

// SetMetadataUint64 set metadata uint64 value
func (p *Packet) SetMetadataUint64(key string, val uint64) {
	if p == nil {
		return
	}
	if p.Metadata == nil {
		p.Metadata = make(map[string]string)
	}
	p.Metadata[key] = strconv.FormatUint(val, 10)
	return
}

// SetMetadataInt64 set metadata int64 value
func (p *Packet) SetMetadataInt64(key string, val int64) {
	if p == nil {
		return
	}
	if p.Metadata == nil {
		p.Metadata = make(map[string]string)
	}
	p.Metadata[key] = strconv.FormatInt(val, 10)
	return
}

func (x *ErrorResponse) MarshalLogObject(e zapcore.ObjectEncoder) (err error) {
	e.AddInt64("code", x.Code)
	e.AddBool("flag", x.LogicError)
	e.AddString("desc", x.Desc)
	return
}

// Error wrap error interface
func (err *ErrorResponse) Error() string {
	return fmt.Sprintf("code:%d logic:%t desc:%s",
		err.GetCode(), err.GetLogicError(),
		err.GetDesc(),
	)
}

// WrapError wrap error to another error
func (err *ErrorResponse) WrapError(in error) (out error) {
	if in == nil {
		return
	}
	out = &ErrorResponse{
		Code:       err.Code,
		LogicError: err.LogicError,
		Desc:       fmt.Sprintf("%s [%v]", err.Desc, in),
	}
	return
}

func (err *ErrorResponse) Equal(in *ErrorResponse) bool {
	if in == nil || err == nil {
		return false
	}
	return err.Code == in.Code
}

func (err *ErrorResponse) EqualError(in error) bool {
	if in == nil || err == nil {
		return false
	}
	if src, ok := in.(*ErrorResponse); ok {
		return err.Code == src.Code
	}
	return false
}

// NewError new error
func NewError(code int64, desc string) (err *ErrorResponse) {
	return &ErrorResponse{
		Code:       code,
		LogicError: true,
		Desc:       desc,
	}
}

func NewInternalError(code ErrorCode, desc string) (err *ErrorResponse) {
	return &ErrorResponse{
		Code:       int64(code),
		LogicError: false,
		Desc:       desc,
	}
}

// Internal Error
var (
	ErrUnkown          = NewInternalError(ErrorCode_UnkownErr, "unkown error")
	ErrMarshalFailed   = NewInternalError(ErrorCode_MarshalFailed, "marshal msg failed")
	ErrUnmarshalFailed = NewInternalError(ErrorCode_UnmarshalFailed, "unmarshal message failed")
	ErrNotSupport      = NewInternalError(ErrorCode_NotSupport, "not support interface")
	ErrTimeout         = NewInternalError(ErrorCode_Timeout, "request timeout")
	ErrPacketTooLarge  = NewInternalError(ErrorCode_PacketTooLarge, "packet size too larse")
	ErrUnexpectedCode  = NewInternalError(ErrorCode_UnexpectedCode, "unexcepted code.check code")
	ErrSessionClosed   = NewInternalError(ErrorCode_SessionClosed, "session closed")
)
