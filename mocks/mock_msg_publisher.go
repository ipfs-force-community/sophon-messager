// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/filecoin-project/venus-messager/publisher (interfaces: IMsgPublisher)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	types "github.com/filecoin-project/venus/venus-shared/types"
	gomock "github.com/golang/mock/gomock"
)

// MockIMsgPublisher is a mock of IMsgPublisher interface.
type MockIMsgPublisher struct {
	ctrl     *gomock.Controller
	recorder *MockIMsgPublisherMockRecorder
}

// MockIMsgPublisherMockRecorder is the mock recorder for MockIMsgPublisher.
type MockIMsgPublisherMockRecorder struct {
	mock *MockIMsgPublisher
}

// NewMockIMsgPublisher creates a new mock instance.
func NewMockIMsgPublisher(ctrl *gomock.Controller) *MockIMsgPublisher {
	mock := &MockIMsgPublisher{ctrl: ctrl}
	mock.recorder = &MockIMsgPublisherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIMsgPublisher) EXPECT() *MockIMsgPublisherMockRecorder {
	return m.recorder
}

// PublishMessages mocks base method.
func (m *MockIMsgPublisher) PublishMessages(arg0 context.Context, arg1 []*types.SignedMessage) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PublishMessages", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// PublishMessages indicates an expected call of PublishMessages.
func (mr *MockIMsgPublisherMockRecorder) PublishMessages(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PublishMessages", reflect.TypeOf((*MockIMsgPublisher)(nil).PublishMessages), arg0, arg1)
}