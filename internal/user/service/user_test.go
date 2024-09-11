package service

import (
	"context"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"oj/internal/middleware"
	"oj/internal/user/domain"
	"oj/internal/user/repository"
	repomocks "oj/internal/user/repository/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserSvc_Login(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (repository.UserRepository, middleware.TokenGenerator)

		// 输入
		identifier string
		password   string
		isEmail    bool

		// 输出
		TokenExpect string
		ErrExpect   error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) (repository.UserRepository, middleware.TokenGenerator) {
				repo := repomocks.NewMockUserRepository(ctrl)
				jwtSvc := middleware.NewMockTokenGenerator(ctrl)
				repo.EXPECT().FindByName(gomock.Any(), "crazyfrank").Return(domain.User{
					Id:       123,
					Password: "$2a$10$Xw4ie5xVmO3OdAxaxEN/NO5VEOKxh9J/Kd37vv02RPKTUZPHwTaBC",
					Role:     0,
				}, nil)
				jwtSvc.EXPECT().GenerateToken(uint8(0), uint64(123), "Mozilla/5.0").Return("mocked-token", nil)
				return repo, jwtSvc
			},
			identifier:  "crazyfrank",
			password:    "hello#12345",
			isEmail:     false,
			TokenExpect: "mocked-token",
			ErrExpect:   nil,
		},
		{
			name: "用户不存在",
			mock: func(ctrl *gomock.Controller) (repository.UserRepository, middleware.TokenGenerator) {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByName(gomock.Any(), "crazyfrank").Return(domain.User{}, repository.ErrUserNotFound)
				return repo, nil
			},
			identifier: "crazyfrank",
			password:   "hello#12345",
			isEmail:    false,
			ErrExpect:  repository.ErrUserNotFound,
		},
		{
			name: "密码不正确",
			mock: func(ctrl *gomock.Controller) (repository.UserRepository, middleware.TokenGenerator) {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByName(gomock.Any(), "crazyfrank").Return(domain.User{
					Id:       123,
					Password: "$2a$10$Xw4ie5xVmO3OdAxaxEN/NO5VEOKxh9J/Kd37vv02RPKTUZPHwTaBC",
					Role:     0,
				}, nil)
				return repo, nil
			},
			identifier: "crazyfrank",
			password:   "hell#1234",
			isEmail:    false,
			ErrExpect:  ErrInvalidUserOrPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewUserService(tc.mock(ctrl))

			ctx := context.WithValue(context.Background(), "UserAgent", "Mozilla/5.0")
			token, err := svc.Login(ctx, tc.identifier, tc.password, tc.isEmail)

			assert.Equal(t, tc.TokenExpect, token)
			assert.Equal(t, tc.ErrExpect, err)
		})
	}
}

func TestUserSvc_EditInfo(t *testing.T) {

}

func TestUserSvc_FindOrCreate(t *testing.T) {

}

func TestUserSvc_GenerateCode(t *testing.T) {

}

func TestUserSvc_GetInfo(t *testing.T) {

}

func TestUserSvc_Signup(t *testing.T) {

}

func TestEncrypted(t *testing.T) {
	res, err := bcrypt.GenerateFromPassword([]byte("hello#12345"), bcrypt.DefaultCost)
	if err == nil {
		t.Log(string(res))
	}
}
