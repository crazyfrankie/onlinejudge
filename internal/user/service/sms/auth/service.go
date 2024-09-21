package auth

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt"
	"oj/internal/user/service/sms"
)

type SMSService struct {
	svc sms.Service
	key string
}

type Claims struct {
	jwt.StandardClaims
	// 真正要用的 tplId
	tplId string
}

// Send 发送，其中 biz 必须是线下申请的一个代表业务方的 token
func (s *SMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	var tc Claims

	// 在这个地方解析
	// 如果解析成功，代表就是对应的业务方
	token, err := jwt.ParseWithClaims(biz, tc, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("token is invalid")
	}
	return s.Send(ctx, tc.tplId, args, numbers...)
}
