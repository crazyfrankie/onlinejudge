package web

import (
	"bytes"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"oj/internal/middleware"
	"testing"

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

			hdl := NewUserHandler(tc.mock(ctrl), nil, nil)
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
			hdl := NewUserHandler(tc.mock(ctrl), nil, nil)
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

func TestUserHandler_SignupSendSMSCode(t *testing.T) {
	testCases := []struct {
		name       string
		req        string
		expectCode int
		expectResp string
		mock       func(ctrl *gomock.Controller) service.CodeService
	}{
		{
			name:       "发送成功",
			req:        `phone=13012345678`,
			expectCode: http.StatusOK,
			expectResp: "\"send successfully\"",
			mock: func(ctrl *gomock.Controller) service.CodeService {
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Send(gomock.Any(), "signup", "13012345678").
					Return(nil)
				return codeSvc
			},
		},
		{
			name:       "发送次数太多",
			req:        `phone=13012345678`,
			expectCode: http.StatusTooManyRequests,
			expectResp: "\"send too many\"",
			mock: func(ctrl *gomock.Controller) service.CodeService {
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Send(gomock.Any(), "signup", "13012345678").
					Return(service.ErrSendTooMany)
				return codeSvc
			},
		},
		{
			name:       "系统错误",
			req:        `phone=13012345678`,
			expectCode: http.StatusInternalServerError,
			expectResp: "\"system error\"",
			mock: func(ctrl *gomock.Controller) service.CodeService {
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Send(gomock.Any(), "signup", "13012345678").
					Return(errors.New("internal error"))
				return codeSvc
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			hdl := NewUserHandler(nil, tc.mock(ctrl), nil)
			hdl.RegisterRoute(server)

			req, err := http.NewRequest(http.MethodPost, "/user/signup/send-code", bytes.NewBuffer([]byte(tc.req)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.expectCode, resp.Code)
			assert.Equal(t, tc.expectResp, resp.Body.String())
		})
	}
}

func TestUserHandler_SignupVerifySMSCode(t *testing.T) {
	testCases := []struct {
		name       string
		req        string
		expectCode int
		expectResp string
		mock       func(ctrl *gomock.Controller) service.CodeService
	}{
		{
			name:       "验证成功",
			req:        `phone=13012345678&code=1234`,
			expectCode: http.StatusOK,
			expectResp: "\"verification successfully\"",
			mock: func(ctrl *gomock.Controller) service.CodeService {
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), "signup", "13012345678", "1234").
					Return(true, nil)
				return codeSvc
			},
		},
		{
			name:       "验证次数频繁",
			req:        `phone=13012345678&code=1234`,
			expectCode: http.StatusBadRequest,
			expectResp: "\"too many verifications\"",
			mock: func(ctrl *gomock.Controller) service.CodeService {
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), "signup", "13012345678", "1234").
					Return(false, service.ErrVerifyTooMany)
				return codeSvc
			},
		},
		{
			name:       "系统错误",
			req:        `phone=13012345678&code=1234`,
			expectCode: http.StatusInternalServerError,
			expectResp: "\"system error\"",
			mock: func(ctrl *gomock.Controller) service.CodeService {
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), "signup", "13012345678", "1234").
					Return(false, errors.New("internal server error"))
				return codeSvc
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			hdl := NewUserHandler(nil, tc.mock(ctrl), nil)
			hdl.RegisterRoute(server)

			req, err := http.NewRequest(http.MethodPost, "/user/signup/verify-code", bytes.NewBuffer([]byte(tc.req)))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.expectCode, resp.Code)
			assert.Equal(t, tc.expectResp, resp.Body.String())
		})

	}
}

func TestUserHandler_LoginSendSMSCode(t *testing.T) {
	testCases := []struct {
		name       string
		req        string
		expectCode int
		expectResp string
		mock       func(ctrl *gomock.Controller) service.CodeService
	}{
		{
			name: "发送成功",
			req: `{
					"phone":"13012345678"
				}`,
			expectCode: http.StatusOK,
			expectResp: "\"send successfully\"",
			mock: func(ctrl *gomock.Controller) service.CodeService {
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Send(gomock.Any(), "login", "13012345678").
					Return(nil)
				return codeSvc
			},
		},
		{
			name: "手机号格式错误",
			req: `{
					"phone":"12312345678"
				}`,
			expectCode: http.StatusBadRequest,
			expectResp: "\"your phone number does not fit the format\"",
			mock: func(ctrl *gomock.Controller) service.CodeService {
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return codeSvc
			},
		},
		{
			name: "发送次数过多",
			req: `{
					"phone":"13012345678"
				}`,
			expectCode: http.StatusTooManyRequests,
			expectResp: "\"send too many\"",
			mock: func(ctrl *gomock.Controller) service.CodeService {
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Send(gomock.Any(), "login", "13012345678").
					Return(service.ErrSendTooMany)
				return codeSvc
			},
		},
		{
			name: "系统错误",
			req: `{
					"phone":"13012345678"
				}`,
			expectCode: http.StatusInternalServerError,
			expectResp: "\"system error\"",
			mock: func(ctrl *gomock.Controller) service.CodeService {
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Send(gomock.Any(), "login", "13012345678").
					Return(errors.New("internal server error"))
				return codeSvc
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			hdl := NewUserHandler(nil, tc.mock(ctrl), nil)
			hdl.RegisterRoute(server)

			req, err := http.NewRequest(http.MethodPost, "/user/login/send-code", bytes.NewBuffer([]byte(tc.req)))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.expectCode, resp.Code)
			assert.Equal(t, tc.expectResp, resp.Body.String())
		})
	}
}

func TestUserHandler_LoginVerifySMSCode(t *testing.T) {
	testCases := []struct {
		name        string
		req         string
		expectCode  int
		expectResp  string
		expectToken string
		mock        func(ctrl *gomock.Controller) (service.UserService, service.CodeService, middleware.TokenGenerator)
	}{
		{
			name: "登录成功",
			req: `{
					"phone":"13012345678",
					"code":"123456"
				}`,
			expectCode:  http.StatusOK,
			expectResp:  "\"login successfully\"",
			expectToken: "valid_jwt_token",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, middleware.TokenGenerator) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().FindOrCreate(gomock.Any(), "13012345678").
					Return(domain.User{Id: uint64(123456), Role: uint8(0)}, nil)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), "login", "13012345678", "123456").
					Return(true, nil)

				tokenGen := middleware.NewMockTokenGenerator(ctrl)
				tokenGen.EXPECT().GenerateToken(uint8(0), uint64(123456), gomock.Any()).
					Return("valid_jwt_token", nil)

				return userSvc, codeSvc, tokenGen
			},
		},
		{
			name: "绑定错误, Bind 失败",
			req: `{
					"phone":"13012345678",
				}`,
			expectCode: http.StatusBadRequest,
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, middleware.TokenGenerator) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc, nil
			},
		},
		{
			name: "验证次数过多",
			req: `{
					"phone":"13012345678",
					"code":"123456"
				}`,
			expectCode: http.StatusBadRequest,
			expectResp: "\"too many verifications\"",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, middleware.TokenGenerator) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), "login", "13012345678", "123456").
					Return(false, service.ErrVerifyTooMany)
				return userSvc, codeSvc, nil
			},
		},
		{
			name: "系统错误 - 验证阶段",
			req: `{
					"phone":"13012345678",
					"code":"123456"
				}`,
			expectCode: http.StatusInternalServerError,
			expectResp: "\"system error\"",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, middleware.TokenGenerator) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), "login", "13012345678", "123456").
					Return(false, errors.New("internal server error"))
				return userSvc, codeSvc, nil
			},
		},
		{
			name: "用户没找到",
			req: `{
					"phone":"13012345678",
					"code":"123456"
				}`,
			expectCode: http.StatusInternalServerError,
			expectResp: "\"identifier not found\"",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, middleware.TokenGenerator) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().FindOrCreate(gomock.Any(), "13012345678").
					Return(domain.User{}, service.ErrUserNotFound)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), "login", "13012345678", "123456").
					Return(true, nil)
				return userSvc, codeSvc, nil
			},
		},
		{
			name: "系统错误 - 用户查找阶段",
			req: `{
					"phone":"13012345678",
					"code":"123456"
				}`,
			expectCode: http.StatusBadRequest,
			expectResp: "\"system error\"",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, middleware.TokenGenerator) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().FindOrCreate(gomock.Any(), "13012345678").
					Return(domain.User{}, errors.New("internal server error"))
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), "login", "13012345678", "123456").
					Return(true, nil)
				return userSvc, codeSvc, nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			hdl := NewUserHandler(tc.mock(ctrl))
			hdl.RegisterRoute(server)

			req, err := http.NewRequest(http.MethodPost, "/user/login-sms", bytes.NewBuffer([]byte(tc.req)))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Agent", "test-agent")

			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.expectCode, resp.Code)
			assert.Equal(t, tc.expectResp, resp.Body.String())

			if tc.expectToken != "" {
				assert.Equal(t, tc.expectToken, resp.Header().Get("x-jwt-token"))
			}
		})
	}
}
