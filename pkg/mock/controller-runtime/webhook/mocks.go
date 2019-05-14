// Code generated by MockGen. DO NOT EDIT.
// Source: sigs.k8s.io/controller-runtime/pkg/webhook (interfaces: Webhook)

// Package webhook is a generated GoMock package.
package webhook

import (
	gomock "github.com/golang/mock/gomock"
	http "net/http"
	reflect "reflect"
	types "sigs.k8s.io/controller-runtime/pkg/webhook/types"
)

// MockWebhook is a mock of Webhook interface
type MockWebhook struct {
	ctrl     *gomock.Controller
	recorder *MockWebhookMockRecorder
}

// MockWebhookMockRecorder is the mock recorder for MockWebhook
type MockWebhookMockRecorder struct {
	mock *MockWebhook
}

// NewMockWebhook creates a new mock instance
func NewMockWebhook(ctrl *gomock.Controller) *MockWebhook {
	mock := &MockWebhook{ctrl: ctrl}
	mock.recorder = &MockWebhookMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockWebhook) EXPECT() *MockWebhookMockRecorder {
	return m.recorder
}

// GetName mocks base method
func (m *MockWebhook) GetName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetName indicates an expected call of GetName
func (mr *MockWebhookMockRecorder) GetName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetName", reflect.TypeOf((*MockWebhook)(nil).GetName))
}

// GetPath mocks base method
func (m *MockWebhook) GetPath() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPath")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetPath indicates an expected call of GetPath
func (mr *MockWebhookMockRecorder) GetPath() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPath", reflect.TypeOf((*MockWebhook)(nil).GetPath))
}

// GetType mocks base method
func (m *MockWebhook) GetType() types.WebhookType {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetType")
	ret0, _ := ret[0].(types.WebhookType)
	return ret0
}

// GetType indicates an expected call of GetType
func (mr *MockWebhookMockRecorder) GetType() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetType", reflect.TypeOf((*MockWebhook)(nil).GetType))
}

// Handler mocks base method
func (m *MockWebhook) Handler() http.Handler {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Handler")
	ret0, _ := ret[0].(http.Handler)
	return ret0
}

// Handler indicates an expected call of Handler
func (mr *MockWebhookMockRecorder) Handler() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Handler", reflect.TypeOf((*MockWebhook)(nil).Handler))
}

// Validate mocks base method
func (m *MockWebhook) Validate() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Validate")
	ret0, _ := ret[0].(error)
	return ret0
}

// Validate indicates an expected call of Validate
func (mr *MockWebhookMockRecorder) Validate() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Validate", reflect.TypeOf((*MockWebhook)(nil).Validate))
}
