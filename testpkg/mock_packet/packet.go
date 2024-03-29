// Code generated by MockGen. DO NOT EDIT.
// Source: packet.go

// Package mock_packet is a generated GoMock package.
package mock_packet

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	message "github.com/walleframe/walle/process/message"
	metadata "github.com/walleframe/walle/process/metadata"
	packet "github.com/walleframe/walle/process/packet"
)

// MockEncoder is a mock of Encoder interface.
type MockEncoder struct {
	ctrl     *gomock.Controller
	recorder *MockEncoderMockRecorder
}

// MockEncoderMockRecorder is the mock recorder for MockEncoder.
type MockEncoderMockRecorder struct {
	mock *MockEncoder
}

// NewMockEncoder creates a new mock instance.
func NewMockEncoder(ctrl *gomock.Controller) *MockEncoder {
	mock := &MockEncoder{ctrl: ctrl}
	mock.recorder = &MockEncoderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEncoder) EXPECT() *MockEncoderMockRecorder {
	return m.recorder
}

// Decode mocks base method.
func (m *MockEncoder) Decode(buf []byte) []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Decode", buf)
	ret0, _ := ret[0].([]byte)
	return ret0
}

// Decode indicates an expected call of Decode.
func (mr *MockEncoderMockRecorder) Decode(buf interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Decode", reflect.TypeOf((*MockEncoder)(nil).Decode), buf)
}

// Encode mocks base method.
func (m *MockEncoder) Encode(buf []byte) []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Encode", buf)
	ret0, _ := ret[0].([]byte)
	return ret0
}

// Encode indicates an expected call of Encode.
func (mr *MockEncoderMockRecorder) Encode(buf interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Encode", reflect.TypeOf((*MockEncoder)(nil).Encode), buf)
}

// MockPool is a mock of Pool interface.
type MockPool struct {
	ctrl     *gomock.Controller
	recorder *MockPoolMockRecorder
}

// MockPoolMockRecorder is the mock recorder for MockPool.
type MockPoolMockRecorder struct {
	mock *MockPool
}

// NewMockPool creates a new mock instance.
func NewMockPool(ctrl *gomock.Controller) *MockPool {
	mock := &MockPool{ctrl: ctrl}
	mock.recorder = &MockPoolMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPool) EXPECT() *MockPoolMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockPool) Get() interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get")
	ret0, _ := ret[0].(interface{})
	return ret0
}

// Get indicates an expected call of Get.
func (mr *MockPoolMockRecorder) Get() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockPool)(nil).Get))
}

// Put mocks base method.
func (m *MockPool) Put(arg0 interface{}) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Put", arg0)
}

// Put indicates an expected call of Put.
func (mr *MockPoolMockRecorder) Put(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Put", reflect.TypeOf((*MockPool)(nil).Put), arg0)
}

// MockCodec is a mock of Codec interface.
type MockCodec struct {
	ctrl     *gomock.Controller
	recorder *MockCodecMockRecorder
}

// MockCodecMockRecorder is the mock recorder for MockCodec.
type MockCodecMockRecorder struct {
	mock *MockCodec
}

// NewMockCodec creates a new mock instance.
func NewMockCodec(ctrl *gomock.Controller) *MockCodec {
	mock := &MockCodec{ctrl: ctrl}
	mock.recorder = &MockCodecMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCodec) EXPECT() *MockCodecMockRecorder {
	return m.recorder
}

// Marshal mocks base method.
func (m *MockCodec) Marshal(p interface{}) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Marshal", p)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Marshal indicates an expected call of Marshal.
func (mr *MockCodecMockRecorder) Marshal(p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Marshal", reflect.TypeOf((*MockCodec)(nil).Marshal), p)
}

// Unmarshal mocks base method.
func (m *MockCodec) Unmarshal(data []byte, p interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Unmarshal", data, p)
	ret0, _ := ret[0].(error)
	return ret0
}

// Unmarshal indicates an expected call of Unmarshal.
func (mr *MockCodecMockRecorder) Unmarshal(data, p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unmarshal", reflect.TypeOf((*MockCodec)(nil).Unmarshal), data, p)
}

// MockProtocolWraper is a mock of ProtocolWraper interface.
type MockProtocolWraper struct {
	ctrl     *gomock.Controller
	recorder *MockProtocolWraperMockRecorder
}

// MockProtocolWraperMockRecorder is the mock recorder for MockProtocolWraper.
type MockProtocolWraperMockRecorder struct {
	mock *MockProtocolWraper
}

// NewMockProtocolWraper creates a new mock instance.
func NewMockProtocolWraper(ctrl *gomock.Controller) *MockProtocolWraper {
	mock := &MockProtocolWraper{ctrl: ctrl}
	mock.recorder = &MockProtocolWraperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProtocolWraper) EXPECT() *MockProtocolWraperMockRecorder {
	return m.recorder
}

// GetMetadata mocks base method.
func (m *MockProtocolWraper) GetMetadata(pkg interface{}) (metadata.MD, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMetadata", pkg)
	ret0, _ := ret[0].(metadata.MD)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMetadata indicates an expected call of GetMetadata.
func (mr *MockProtocolWraperMockRecorder) GetMetadata(pkg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMetadata", reflect.TypeOf((*MockProtocolWraper)(nil).GetMetadata), pkg)
}

// NewPacket mocks base method.
func (m *MockProtocolWraper) NewPacket(inPkg interface{}, cmd packet.PacketCmd, uri interface{}, md metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewPacket", inPkg, cmd, uri, md)
	ret0, _ := ret[0].(error)
	return ret0
}

// NewPacket indicates an expected call of NewPacket.
func (mr *MockProtocolWraperMockRecorder) NewPacket(inPkg, cmd, uri, md interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewPacket", reflect.TypeOf((*MockProtocolWraper)(nil).NewPacket), inPkg, cmd, uri, md)
}

// NewResponse mocks base method.
func (m *MockProtocolWraper) NewResponse(inPkg, outPkg interface{}, md metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewResponse", inPkg, outPkg, md)
	ret0, _ := ret[0].(error)
	return ret0
}

// NewResponse indicates an expected call of NewResponse.
func (mr *MockProtocolWraperMockRecorder) NewResponse(inPkg, outPkg, md interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewResponse", reflect.TypeOf((*MockProtocolWraper)(nil).NewResponse), inPkg, outPkg, md)
}

// PayloadMarshal mocks base method.
func (m *MockProtocolWraper) PayloadMarshal(pkg interface{}, codec message.Codec, payload interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PayloadMarshal", pkg, codec, payload)
	ret0, _ := ret[0].(error)
	return ret0
}

// PayloadMarshal indicates an expected call of PayloadMarshal.
func (mr *MockProtocolWraperMockRecorder) PayloadMarshal(pkg, codec, payload interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PayloadMarshal", reflect.TypeOf((*MockProtocolWraper)(nil).PayloadMarshal), pkg, codec, payload)
}

// PayloadUnmarshal mocks base method.
func (m *MockProtocolWraper) PayloadUnmarshal(pkg interface{}, codec message.Codec, obj interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PayloadUnmarshal", pkg, codec, obj)
	ret0, _ := ret[0].(error)
	return ret0
}

// PayloadUnmarshal indicates an expected call of PayloadUnmarshal.
func (mr *MockProtocolWraperMockRecorder) PayloadUnmarshal(pkg, codec, obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PayloadUnmarshal", reflect.TypeOf((*MockProtocolWraper)(nil).PayloadUnmarshal), pkg, codec, obj)
}
