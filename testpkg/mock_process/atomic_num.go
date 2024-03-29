// Code generated by MockGen. DO NOT EDIT.
// Source: atomic_num.go

// Package mock_process is a generated GoMock package.
package mock_process

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockAtomicNumber is a mock of AtomicNumber interface.
type MockAtomicNumber struct {
	ctrl     *gomock.Controller
	recorder *MockAtomicNumberMockRecorder
}

// MockAtomicNumberMockRecorder is the mock recorder for MockAtomicNumber.
type MockAtomicNumberMockRecorder struct {
	mock *MockAtomicNumber
}

// NewMockAtomicNumber creates a new mock instance.
func NewMockAtomicNumber(ctrl *gomock.Controller) *MockAtomicNumber {
	mock := &MockAtomicNumber{ctrl: ctrl}
	mock.recorder = &MockAtomicNumberMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAtomicNumber) EXPECT() *MockAtomicNumberMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockAtomicNumber) Add(n int64) int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", n)
	ret0, _ := ret[0].(int64)
	return ret0
}

// Add indicates an expected call of Add.
func (mr *MockAtomicNumberMockRecorder) Add(n interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockAtomicNumber)(nil).Add), n)
}

// Dec mocks base method.
func (m *MockAtomicNumber) Dec() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Dec")
	ret0, _ := ret[0].(int64)
	return ret0
}

// Dec indicates an expected call of Dec.
func (mr *MockAtomicNumberMockRecorder) Dec() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Dec", reflect.TypeOf((*MockAtomicNumber)(nil).Dec))
}

// Inc mocks base method.
func (m *MockAtomicNumber) Inc() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Inc")
	ret0, _ := ret[0].(int64)
	return ret0
}

// Inc indicates an expected call of Inc.
func (mr *MockAtomicNumberMockRecorder) Inc() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Inc", reflect.TypeOf((*MockAtomicNumber)(nil).Inc))
}

// Load mocks base method.
func (m *MockAtomicNumber) Load() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Load")
	ret0, _ := ret[0].(int64)
	return ret0
}

// Load indicates an expected call of Load.
func (mr *MockAtomicNumberMockRecorder) Load() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Load", reflect.TypeOf((*MockAtomicNumber)(nil).Load))
}

// Store mocks base method.
func (m *MockAtomicNumber) Store(n int64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Store", n)
}

// Store indicates an expected call of Store.
func (mr *MockAtomicNumberMockRecorder) Store(n interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Store", reflect.TypeOf((*MockAtomicNumber)(nil).Store), n)
}

// Sub mocks base method.
func (m *MockAtomicNumber) Sub(n int64) int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Sub", n)
	ret0, _ := ret[0].(int64)
	return ret0
}

// Sub indicates an expected call of Sub.
func (mr *MockAtomicNumberMockRecorder) Sub(n interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sub", reflect.TypeOf((*MockAtomicNumber)(nil).Sub), n)
}
