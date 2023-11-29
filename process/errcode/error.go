package errcode

import (
	"encoding/binary"
	"fmt"

	"github.com/walleframe/walle/util"
)

// ErrorResponse represent rpc call common error
type ErrorResponse struct {
	// error code
	Code uint32
	// desc
	Desc string
}

// Error implement error interface.
func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%d-%v", e.Code, e.Desc)
}

// WrapError wrap error to another error
func (err *ErrorResponse) WrapError(in error) (out error) {
	if in == nil {
		return
	}
	out = &ErrorResponse{
		Code: err.Code,
		Desc: fmt.Sprintf("%s [%v]", err.Desc, in),
	}
	return
}

func (err *ErrorResponse) Codes() uint32 {
	return err.Code
}

var _ error = (*ErrorResponse)(nil)

func NewError(code ErrorCode, desc string) error {
	return &ErrorResponse{
		Code: uint32(code),
		Desc: desc,
	}
}

func Is(err error, code ErrorCode) bool {
	if c, ok := err.(*ErrorResponse); ok {
		return c.Code == uint32(code)
	}
	return false
}

// WrapError wrap error to another error
var WrapError = func(code, other error) (out error) {
	if other == nil {
		return nil
	}
	if c, ok := code.(*ErrorResponse); ok {
		return c.WrapError(other)
	}
	return &ErrorResponse{
		Code: uint32(ErrorCodeUnkwon),
		Desc: fmt.Sprintf("%v [%v]", code, other),
	}
}

type errResponseCodec struct {
}

func (errResponseCodec) Marshal(code error) (data []byte, err error) {
	e, ok := code.(*ErrorResponse)
	if !ok {
		tip := code.Error()
		data = make([]byte, 4+len(tip)) // mempool.Pool().Alloc(4 + len(tip))
		// code default is 1,unkown error
		binary.BigEndian.PutUint32(data, 1)
		data = append(data, util.StringToBytes(tip)...)
		return
	}
	data = make([]byte, 4+len(e.Desc)) // mempool.Pool().Alloc(4 + len(e.Desc))
	binary.BigEndian.PutUint32(data, e.Code)
	data = append(data, util.StringToBytes(e.Desc)...)
	return
}

func (errResponseCodec) Unmarshal(data []byte) (err error) {
	if len(data) < 4 {
		return ErrPacketsizeInvalid
	}
	e := &ErrorResponse{}
	e.Code = binary.BigEndian.Uint32(data)
	e.Desc = string(data[4:])
	return e
}

type errResponseNew struct{}

func (errResponseNew) New() error {
	return &ErrorResponse{}
}
