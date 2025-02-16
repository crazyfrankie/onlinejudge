// Code generated by MockGen. DO NOT EDIT.
// Source: D:\Gocode\oj\internal\user\repository\cache\user.go

// Package cachemocks is a generated GoMock package.
package cachemocks

import (
	context "context"
	domain "github.com/crazyfrankie/onlinejudge/internal/user/domain"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockUserCache is a mock of UserCache interface.
type MockUserCache struct {
	ctrl     *gomock.Controller
	recorder *MockUserCacheMockRecorder
}

// MockUserCacheMockRecorder is the mock recorder for MockUserCache.
type MockUserCacheMockRecorder struct {
	mock *MockUserCache
}

// NewMockUserCache creates a new mock instance.
func NewMockUserCache(ctrl *gomock.Controller) *MockUserCache {
	mock := &MockUserCache{ctrl: ctrl}
	mock.recorder = &MockUserCacheMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserCache) EXPECT() *MockUserCacheMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockUserCache) Get(ctx context.Context, id uint64) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, id)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockUserCacheMockRecorder) Get(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockUserCache)(nil).Get), ctx, id)
}

// Set mocks base method.
func (m *MockUserCache) Set(ctx context.Context, user domain.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", ctx, user)
	ret0, _ := ret[0].(error)
	return ret0
}

// Set indicates an expected call of Set.
func (mr *MockUserCacheMockRecorder) Set(ctx, user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockUserCache)(nil).Set), ctx, user)
}

// key mocks base method.
func (m *MockUserCache) key(id uint64) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "key", id)
	ret0, _ := ret[0].(string)
	return ret0
}

// key indicates an expected call of key.
func (mr *MockUserCacheMockRecorder) key(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "key", reflect.TypeOf((*MockUserCache)(nil).key), id)
}
