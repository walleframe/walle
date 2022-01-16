package errcode

// ErrorCodec error codec for custom error define
type ErrorCodec interface {
	// Marshal marshal errCode to binary data
	Marshal(errCode error) (data []byte, err error)
	// Unmarshal unmarshal errCode from binary data
	Unmarshal(data []byte) (err error)
}

// DefaultErrorCodec default error codec
var DefaultErrorCodec ErrorCodec = errResponseCodec{}

// ErrorsNew use for new custom error
type ErrorsNew interface {
	New() error
}

// DefaultErrorsNew default new error funcs
var DefaultErrorsNew = errResponseNew{}
