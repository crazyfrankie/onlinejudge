package web

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"oj/internal/user/domain"
	"oj/internal/user/service"
	svcmocks "oj/internal/user/service/mocks"
)

func TestUserHandler_Signup(t *testing.T) {
	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) service.UserService
		reqBody    string
		expectCode int
		expectResp string
	}{
		{
			name: "注册成功",
			reqBody: `{
				"name":"crazybruce",
				"password":"hello#12345",
				"confirmPassword":"hello#12345",
				"email":"123456@163.com",
				"phone":"13012345678",
				"role":0
			}`,
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{
					Name:     "crazybruce",
					Password: "hello#12345",
					Email:    "123456@163.com",
					Phone:    "13012345678",
					Role:     0,
				}).Return(nil)
				return userSvc
			},
			expectCode: http.StatusOK,
			expectResp: "\"sign up successfully!\"",
		},
		{
			name: "参数格式不对, bind 失败",
			reqBody: `{
				"name":"crazybruce",
				"password":"hello#12345",
				"confirmPassword":"hello#12345",
				"email":"123456@163.com",
				"phone":"13012345678",
			}`,
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc
			},
			expectCode: http.StatusBadRequest,
		},
		{
			name: "两次密码不一致",
			reqBody: `{
				"name":"crazybruce",
				"password":"hello#12345",
				"confirmPassword":"hello#1234",
				"email":"123456@163.com",
				"phone":"13012345678",
				"role":0
			}`,
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc
			},
			expectCode: http.StatusBadRequest,
			expectResp: "\"password does not match\"",
		},
		{
			name: "邮箱格式不对",
			reqBody: `{
				"name":"crazybruce",
				"password":"hello#12345",
				"confirmPassword":"hello#12345",
				"email":"123456@qq",
				"phone":"13012345678",
				"role":0
			}`,
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc
			},
			expectCode: http.StatusBadRequest,
			expectResp: "\"your email does not fit the format\"",
		},
		{
			name: "密码格式不对",
			reqBody: `{
				"name":"crazybruce",
				"password":"hello12345",
				"confirmPassword":"hello12345",
				"email":"123456@qq.com",
				"phone":"13012345678",
				"role":0
			}`,
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc
			},
			expectCode: http.StatusBadRequest,
			expectResp: "\"your password does not fit the format\"",
		},
		{
			name: "手机号格式不对",
			reqBody: `{
				"name":"crazybruce",
				"password":"hello#12345",
				"confirmPassword":"hello#12345",
				"email":"123456@qq.com",
				"phone":"12312345678",
				"role":0
			}`,
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc
			},
			expectCode: http.StatusBadRequest,
			expectResp: "\"your phone number does not fit the format\"",
		},
		{
			name: "邮箱冲突",
			reqBody: `{
				"name":"crazybruce",
				"password":"hello#12345",
				"confirmPassword":"hello#12345",
				"email":"123456@163.com",
				"phone":"13012345678",
				"role":0
			}`,
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{
					Name:     "crazybruce",
					Password: "hello#12345",
					Email:    "123456@163.com",
					Phone:    "13012345678",
					Role:     0,
				}).Return(service.ErrUserDuplicateEmail)
				return userSvc
			},
			expectCode: http.StatusInternalServerError,
			expectResp: "\"email conflict\"",
		},
		{
			name: "用户名冲突",
			reqBody: `{
				"name":"crazybruce",
				"password":"hello#12345",
				"confirmPassword":"hello#12345",
				"email":"123456@163.com",
				"phone":"13012345678",
				"role":0
			}`,
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{
					Name:     "crazybruce",
					Password: "hello#12345",
					Email:    "123456@163.com",
					Phone:    "13012345678",
					Role:     0,
				}).Return(service.ErrUserDuplicateName)
				return userSvc
			},
			expectCode: http.StatusInternalServerError,
			expectResp: "\"name conflict\"",
		},
		{
			name: "手机号冲突",
			reqBody: `{
				"name":"crazybruce",
				"password":"hello#12345",
				"confirmPassword":"hello#12345",
				"email":"123456@163.com",
				"phone":"13012345678",
				"role":0
			}`,
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{
					Name:     "crazybruce",
					Password: "hello#12345",
					Email:    "123456@163.com",
					Phone:    "13012345678",
					Role:     0,
				}).Return(service.ErrUserDuplicatePhone)
				return userSvc
			},
			expectCode: http.StatusInternalServerError,
			expectResp: "\"phone conflict\"",
		},
		{
			name: "系统异常",
			reqBody: `{
				"name":"crazybruce",
				"password":"hello#12345",
				"confirmPassword":"hello#12345",
				"email":"123456@163.com",
				"phone":"13012345678",
				"role":0
			}`,
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{
					Name:     "crazybruce",
					Password: "hello#12345",
					Email:    "123456@163.com",
					Phone:    "13012345678",
					Role:     0,
				}).Return(errors.New("随便一个 error"))
				return userSvc
			},
			expectCode: http.StatusBadRequest,
			expectResp: "\"system error\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()

			hdl := NewUserHandler(tc.mock(ctrl), nil)
			hdl.RegisterRoute(server)

			req, err := http.NewRequest(http.MethodPost, "/user/signup", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.expectCode, resp.Code)
			assert.Equal(t, tc.expectResp, resp.Body.String())
		})

	}
}

func TestUserHandler_Login(t *testing.T) {
	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) service.UserService
		req        string
		expectCode int
		expectRes  string
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Login(gomock.Any(), "123456@qq.com", "hello#12345", true).Return("mocked-token", nil)
				return userSvc
			},
			req: `{
				"identifier":"123456@qq.com",
				"password":"hello#12345"
			}`,
			expectCode: http.StatusOK,
			expectRes:  "\"login successfully!\"",
		},
		{
			name: "参数格式错误, bind 失败",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc
			},
			req: `{
				"identifier":"123456@qq.com",
			}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name: "用户名或密码错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Login(gomock.Any(), "123456@qq.com", "hello$12345", true).Return("mock-jwt", service.ErrInvalidUserOrPassword)
				return userSvc
			},
			req: `{
				"identifier":"123456@qq.com",
				"password":"hello$12345"
			}`,
			expectCode: http.StatusInternalServerError,
			expectRes:  "\"identifier or password error\"",
		},
		{
			name: "用户不存在",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Login(gomock.Any(), "123456@qq.com", "hello$12345", true).Return("mock-jwt", service.ErrUserNotFound)
				return userSvc
			},
			req: `{
				"identifier":"123456@qq.com",
				"password":"hello$12345"
			}`,
			expectCode: http.StatusInternalServerError,
			expectRes:  "\"identifier not found\"",
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Login(gomock.Any(), "123456@qq.com", "hello$12345", true).Return("mock-jwt", errors.New("随便错误"))
				return userSvc
			},
			req: `{
				"identifier":"123456@qq.com",
				"password":"hello$12345"
			}`,
			expectCode: http.StatusBadRequest,
			expectRes:  "\"system error\"",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			hdl := NewUserHandler(tc.mock(ctrl), nil)
			hdl.RegisterRoute(server)

			req, err := http.NewRequest(http.MethodPost, "/user/login", bytes.NewBuffer([]byte(tc.req)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.expectCode, resp.Code)
			assert.Equal(t, tc.expectRes, resp.Body.String())
		})
	}
}
