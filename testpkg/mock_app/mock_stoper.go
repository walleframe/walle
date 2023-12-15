// Code generated by MockGen. DO NOT EDIT.
// Source: app.go

// Package mock_app is a generated GoMock package.
package mock_app

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockStoper is a mock of Stoper interface.
type MockStoper struct {
	ctrl     *gomock.Controller
	recorder *MockStoperMockRecorder
}

// MockStoperMockRecorder is the mock recorder for MockStoper.
type MockStoperMockRecorder struct {
	mock *MockStoper
}

// NewMockStoper creates a new mock instance.
func NewMockStoper(ctrl *gomock.Controller) *MockStoper {
	mock := &MockStoper{ctrl: ctrl}
	mock.recorder = &MockStoperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStoper) EXPECT() *MockStoperMockRecorder {
	return m.recorder
}

// GetStopChan mocks base method.
func (m *MockStoper) GetStopChan() <-chan struct{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStopChan")
	ret0, _ := ret[0].(<-chan struct{})
	return ret0
}

// GetStopChan indicates an expected call of GetStopChan.
func (mr *MockStoperMockRecorder) GetStopChan() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStopChan", reflect.TypeOf((*MockStoper)(nil).GetStopChan))
}

// IsStop mocks base method.
func (m *MockStoper) IsStop() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsStop")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsStop indicates an expected call of IsStop.
func (mr *MockStoperMockRecorder) IsStop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsStop", reflect.TypeOf((*MockStoper)(nil).IsStop))
}

// Stop mocks base method.
func (m *MockStoper) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop.
func (mr *MockStoperMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockStoper)(nil).Stop))
}