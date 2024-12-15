package proto

type Message interface {
	MarshalObject() (data []byte, err error)
	MarshalSize() (size int)
	MarshalObjectTo(buf []byte) (data []byte, err error)

	UnmarshalObject(data []byte) (err error)
}

func Marshal(v Message) (data []byte, err error) {
	return v.MarshalObject()
}

func Unmarshal(data []byte, v Message) (err error) {
	return v.UnmarshalObject(data)
}
