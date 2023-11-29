package packet

import (
	"encoding/binary"

	"github.com/walleframe/walle/process/errcode"
	"github.com/walleframe/walle/process/metadata"
	"github.com/walleframe/walle/util"
)

var defaultPacketCodec Codec = BytesURICodec

func GetCodec() Codec {
	return defaultPacketCodec
}

func SetPacketCodec(codec Codec) {
	defaultPacketCodec = codec
}

// 1byte cmd 1byte flag 1byte reserved 1byte msg-len 8byte sessionid 4byte payload-size xbyte-uri  xbyte-payload xbyte-metadata
type codecURI struct{}

var BytesURICodec Codec = codecURI{}

func (codecURI) Marshal(p interface{}) ([]byte, error) {
	pkg, ok := p.(*Packet)
	if !ok {
		return nil, errcode.ErrUnexpectedCode
	}
	md, err := metadata.GetCodec().Marshal(pkg.metadata)
	if err != nil {
		return nil, err
	}
	pkg.msgLen = uint8(len(pkg.msgURI))

	size := 16 + len(pkg.msgURI) + len(pkg.payload) + len(md)
	if size+4 > cap(pkg.cache) {
		pkg.cache = make([]byte, size+4) // mempool.Pool().Alloc(size + 4)
	}
	buf := pkg.cache[:size+4] //mempool.Pool().Alloc(size + 4) // free when packet.Pool.Put
	binary.BigEndian.PutUint32(buf, uint32(size))
	data := buf[4:]
	data[0] = byte(pkg.cmd)
	data[1] = byte(pkg.flag)
	data[2] = pkg.reservd
	data[3] = pkg.msgLen
	binary.BigEndian.PutUint64(data[4:], pkg.sessionID)
	binary.BigEndian.PutUint32(data[12:], uint32(len(pkg.payload)))
	copy(data[16:], util.StringToBytes(pkg.msgURI))
	idx := 16 + len(pkg.msgURI)
	copy(data[idx:], pkg.payload)
	if len(md) > 0 {
		copy(data[idx+len(pkg.payload):], md)
		//mempool.Pool().Free(md)
	}
	// if pkg.cache != nil {
	// 	mempool.Pool().Free(pkg.cache)
	// }
	//pkg.cache = buf
	// //
	// mp.Free(pkg.payload)
	// pkg.payload = nil

	return buf, nil
}
func (codecURI) Unmarshal(data []byte, p interface{}) error {
	if len(data) < 16 {
		return errcode.ErrPacketsizeInvalid
	}
	pkg, ok := p.(*Packet)
	if !ok {
		return errcode.ErrUnexpectedCode
	}
	if size := binary.BigEndian.Uint32(data); size+4 != uint32(len(data)) {
		return errcode.ErrPacketsizeInvalid
	}
	data = data[4:]
	pkg.cmd = PacketCmd(data[0])
	pkg.flag = PacketFlag(data[1])
	pkg.reservd = data[2]
	pkg.msgLen = data[3]
	pkg.sessionID = binary.BigEndian.Uint64(data[4:])
	payloadSize := int(binary.BigEndian.Uint32(data[12:]))
	idx := int(16 + pkg.msgLen)
	pkg.msgURI = string(data[16:idx])
	if payloadSize > cap(pkg.cache) {
		pkg.cache = make([]byte, payloadSize) //mempool.Pool().Alloc(payloadSize)
	}
	pkg.payload = pkg.cache[:payloadSize] //  mempool.Pool().Alloc(payloadSize) // free when packet.Pool.Put
	copy(pkg.payload, data[idx:idx+payloadSize])
	return metadata.GetCodec().Unmarshal(data[idx+payloadSize:], pkg.metadata)
}

// 1byte cmd 1byte flag 1byte reserved 1byte empty 8byte sessionid 4byte payload-size 4byte id  xbyte-payload xbyte-metadata
type codecMID struct{}

var BytesMIDCodec Codec = codecMID{}

func (codecMID) Marshal(p interface{}) ([]byte, error) {
	pkg, ok := p.(*Packet)
	if !ok {
		return nil, errcode.ErrUnexpectedCode
	}
	md, err := metadata.GetCodec().Marshal(pkg.metadata)
	if err != nil {
		return nil, err
	}
	size := 20 + len(pkg.payload) + len(md)
	if size+4 > cap(pkg.cache) {
		pkg.cache = make([]byte, size+4) //.Pool().Alloc(size + 4)
	}
	buf := pkg.cache[:size+4] //buf := mempool.Pool().Alloc(size + 4) // free when packet.Pool.Put
	binary.BigEndian.PutUint32(buf, uint32(size))
	data := buf[4:]
	data[0] = byte(pkg.cmd)
	data[1] = byte(pkg.flag)
	data[2] = pkg.reservd
	//data[3] = emtpy
	binary.BigEndian.PutUint64(data[4:], pkg.sessionID)
	binary.BigEndian.PutUint32(data[12:], uint32(len(pkg.payload)))
	binary.BigEndian.PutUint32(data[16:], pkg.msgID)
	copy(data[20:], pkg.payload)
	if len(md) > 0 {
		copy(data[20+len(pkg.payload):], md)
		//mempool.Pool().Free(md)
	}
	// if pkg.cache != nil {
	// 	mempool.Pool().Free(pkg.cache)
	// }
	//pkg.cache = buf

	return buf, nil
}
func (codecMID) Unmarshal(data []byte, p interface{}) error {
	if len(data) < 20 {
		return errcode.ErrPacketsizeInvalid
	}
	pkg, ok := p.(*Packet)
	if !ok {
		return errcode.ErrUnexpectedCode
	}
	if size := binary.BigEndian.Uint32(data); size+4 != uint32(len(data)) {
		return errcode.ErrPacketsizeInvalid
	}
	data = data[4:]
	pkg.cmd = PacketCmd(data[0])
	pkg.flag = PacketFlag(data[1])
	pkg.reservd = data[2]
	//pkg.msgLen = data[3]
	pkg.sessionID = binary.BigEndian.Uint64(data[4:])
	payloadSize := int(binary.BigEndian.Uint32(data[12:]))
	pkg.msgID = binary.BigEndian.Uint32(data[16:])
	if payloadSize > cap(pkg.cache) {
		pkg.cache = make([]byte, payloadSize) // mempool.Pool().Alloc(payloadSize)
	}
	pkg.payload = pkg.cache[:payloadSize]
	//pkg.payload = mempool.Pool().Alloc(payloadSize) // free when packet.Pool.Put
	copy(pkg.payload, data[20:20+payloadSize])
	return metadata.GetCodec().Unmarshal(data[20+payloadSize:], pkg.metadata)
}
