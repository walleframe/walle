package errcode

// ErrorCode 0:Success 1-1000: frame internal error 1000+: custom logic error
type ErrorCode uint32

const (
	// ErrorCodeSuccess no error
	ErrorCodeSuccess ErrorCode = 0
	// ErrorCodeUnkown unkown error
	ErrorCodeUnkwon ErrorCode = 1
	// marshal msg falied
	ErrorCodeMarshalFailed ErrorCode = 2
	// unmarshal msg failed
	ErrorCodeUnmarshalFailed ErrorCode = 3
	// not support interface,not implemented
	ErrorCodeNotSupport ErrorCode = 4
	// timeout
	ErrorCodeTimeout ErrorCode = 5
	// packet size invalid
	ErrorCodePacketSizeInvalid ErrorCode = 6
	// coding wrong
	ErrorCodeUnexpectedCode ErrorCode = 7
	// session closed
	ErrorCodeSessionClosed ErrorCode = 8
	//
	ErrorCodeInvalidErrorPayload ErrorCode = 9
)

var (
	// ErrorCodeUnkown unkown error
	ErrUnkwon = NewError(ErrorCodeUnkwon, "unkown error")
	// marshal msg falied
	ErrMarshalFailed = NewError(ErrorCodeMarshalFailed, "marshal msg falied")
	// unmarshal msg failed
	ErrUnmarshalFailed = NewError(ErrorCodeUnmarshalFailed, "unmarshal msg failed")
	// not support interface,not implemented
	ErrNotSupport = NewError(ErrorCodeNotSupport, "not support interface,not implemented")
	// timeout
	ErrTimeout = NewError(ErrorCodeTimeout, "timeout")
	// packet size invalid
	ErrPacketsizeInvalid = NewError(ErrorCodePacketSizeInvalid, "packet size too large")
	// coding wrong
	ErrUnexpectedCode = NewError(ErrorCodeUnexpectedCode, "coding wrong")
	// session closed
	ErrSessionClosed = NewError(ErrorCodeSessionClosed, "session closed")
	// ErrInvalidErrPayload error payload invalid
	ErrInvalidErrPayload = NewError(ErrorCodeInvalidErrorPayload, "error payload invalid")
)
