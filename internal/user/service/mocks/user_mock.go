// Code generated by MockGen. DO NOT EDIT.
// Source: D:\Gocode\oj\internal\user\service\user.go

// Package svcmocks is a generated GoMock package.
package svcmocks

import (
	context "context"
	domain "github.com/crazyfrankie/onlinejudge/internal/user/domain"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockUserService is a mock of UserService interface.
type MockUserService struct {
	ctrl     *gomock.Controller
	recorder *MockUserServiceMockRecorder
}

// MockUserServiceMockRecorder is the mock recorder for MockUserService.
type MockUserServiceMockRecorder struct {
	mock *MockUserService
}

// NewMockUserService creates a new mock instance.
func NewMockUserService(ctrl *gomock.Controller) *MockUserService {
	mock := &MockUserService{ctrl: ctrl}
	mock.recorder = &MockUserServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserService) EXPECT() *MockUserServiceMockRecorder {
	return m.recorder
}

// EditInfo mocks base method.
func (m *MockUserService) EditInfo(ctx context.Context, id uint64, user domain.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EditInfo", ctx, id, user)
	ret0, _ := ret[0].(error)
	return ret0
}

// EditInfo indicates an expected call of EditInfo.
func (mr *MockUserServiceMockRecorder) EditInfo(ctx, id, user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EditInfo", reflect.TypeOf((*MockUserService)(nil).EditInfo), ctx, id, user)
}

// FindOrCreate mocks base method.
func (m *MockUserService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindOrCreate", ctx, phone)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindOrCreate indicates an expected call of FindOrCreate.
func (mr *MockUserServiceMockRecorder) FindOrCreate(ctx, phone interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindOrCreate", reflect.TypeOf((*MockUserService)(nil).FindOrCreate), ctx, phone)
}

// GenerateCode mocks base method.
func (m *MockUserService) GenerateCode() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateCode")
	ret0, _ := ret[0].(string)
	return ret0
}

// GenerateCode indicates an expected call of GenerateCode.
func (mr *MockUserServiceMockRecorder) GenerateCode() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateCode", reflect.TypeOf((*MockUserService)(nil).GenerateCode))
}

// GetInfo mocks base method.
func (m *MockUserService) GetInfo(ctx context.Context, id uint64) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInfo", ctx, id)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetInfo indicates an expected call of GetInfo.
func (mr *MockUserServiceMockRecorder) GetInfo(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInfo", reflect.TypeOf((*MockUserService)(nil).GetInfo), ctx, id)
}

// Login mocks base method.
func (m *MockUserService) Login(ctx context.Context, identifier, password string, isEmail bool) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Login", ctx, identifier, password, isEmail)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Login indicates an expected call of Login.
func (mr *MockUserServiceMockRecorder) Login(ctx, identifier, password, isEmail interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Login", reflect.TypeOf((*MockUserService)(nil).Login), ctx, identifier, password, isEmail)
}

// Signup mocks base method.
func (m *MockUserService) Signup(ctx context.Context, u domain.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Signup", ctx, u)
	ret0, _ := ret[0].(error)
	return ret0
}

// Signup indicates an expected call of Signup.
func (mr *MockUserServiceMockRecorder) Signup(ctx, u interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Signup", reflect.TypeOf((*MockUserService)(nil).Signup), ctx, u)
}
