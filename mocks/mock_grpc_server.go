// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/yagoggame/grpc_server (interfaces: Authorizator,Pooler)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	game "github.com/yagoggame/gomaster/game"
	reflect "reflect"
)

// MockAuthorizator is a mock of Authorizator interface
type MockAuthorizator struct {
	ctrl     *gomock.Controller
	recorder *MockAuthorizatorMockRecorder
}

// MockAuthorizatorMockRecorder is the mock recorder for MockAuthorizator
type MockAuthorizatorMockRecorder struct {
	mock *MockAuthorizator
}

// NewMockAuthorizator creates a new mock instance
func NewMockAuthorizator(ctrl *gomock.Controller) *MockAuthorizator {
	mock := &MockAuthorizator{ctrl: ctrl}
	mock.recorder = &MockAuthorizatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAuthorizator) EXPECT() *MockAuthorizatorMockRecorder {
	return m.recorder
}

// Authorize mocks base method
func (m *MockAuthorizator) Authorize(arg0, arg1 string) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Authorize", arg0, arg1)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Authorize indicates an expected call of Authorize
func (mr *MockAuthorizatorMockRecorder) Authorize(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Authorize", reflect.TypeOf((*MockAuthorizator)(nil).Authorize), arg0, arg1)
}

// MockPooler is a mock of Pooler interface
type MockPooler struct {
	ctrl     *gomock.Controller
	recorder *MockPoolerMockRecorder
}

// MockPoolerMockRecorder is the mock recorder for MockPooler
type MockPoolerMockRecorder struct {
	mock *MockPooler
}

// NewMockPooler creates a new mock instance
func NewMockPooler(ctrl *gomock.Controller) *MockPooler {
	mock := &MockPooler{ctrl: ctrl}
	mock.recorder = &MockPoolerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPooler) EXPECT() *MockPoolerMockRecorder {
	return m.recorder
}

// AddGamer mocks base method
func (m *MockPooler) AddGamer(arg0 *game.Gamer) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddGamer", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddGamer indicates an expected call of AddGamer
func (mr *MockPoolerMockRecorder) AddGamer(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddGamer", reflect.TypeOf((*MockPooler)(nil).AddGamer), arg0)
}

// GetGamer mocks base method
func (m *MockPooler) GetGamer(arg0 int) (*game.Gamer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGamer", arg0)
	ret0, _ := ret[0].(*game.Gamer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGamer indicates an expected call of GetGamer
func (mr *MockPoolerMockRecorder) GetGamer(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGamer", reflect.TypeOf((*MockPooler)(nil).GetGamer), arg0)
}

// JoinGame mocks base method
func (m *MockPooler) JoinGame(arg0 int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "JoinGame", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// JoinGame indicates an expected call of JoinGame
func (mr *MockPoolerMockRecorder) JoinGame(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "JoinGame", reflect.TypeOf((*MockPooler)(nil).JoinGame), arg0)
}

// ListGamers mocks base method
func (m *MockPooler) ListGamers() []*game.Gamer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListGamers")
	ret0, _ := ret[0].([]*game.Gamer)
	return ret0
}

// ListGamers indicates an expected call of ListGamers
func (mr *MockPoolerMockRecorder) ListGamers() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListGamers", reflect.TypeOf((*MockPooler)(nil).ListGamers))
}

// Release mocks base method
func (m *MockPooler) Release() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Release")
}

// Release indicates an expected call of Release
func (mr *MockPoolerMockRecorder) Release() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Release", reflect.TypeOf((*MockPooler)(nil).Release))
}

// ReleaseGame mocks base method
func (m *MockPooler) ReleaseGame(arg0 int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReleaseGame", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// ReleaseGame indicates an expected call of ReleaseGame
func (mr *MockPoolerMockRecorder) ReleaseGame(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReleaseGame", reflect.TypeOf((*MockPooler)(nil).ReleaseGame), arg0)
}

// RmGamer mocks base method
func (m *MockPooler) RmGamer(arg0 int) (*game.Gamer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RmGamer", arg0)
	ret0, _ := ret[0].(*game.Gamer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RmGamer indicates an expected call of RmGamer
func (mr *MockPoolerMockRecorder) RmGamer(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RmGamer", reflect.TypeOf((*MockPooler)(nil).RmGamer), arg0)
}
