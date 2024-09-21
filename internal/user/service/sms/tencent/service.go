package tencent

import (
	"context"
	"fmt"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"oj/pkg/ratelimit"
)

type Service struct {
	client   *sms.Client
	appId    *string
	signName *string
	limiter  ratelimit.Limiter
}

func ToPtr(c string) *string {
	return &c
}

func (s *Service) ToStringPtrSlice(strings []string) []*string {
	pointers := make([]*string, len(strings))
	for i, str := range strings {
		pointers[i] = &str
	}
	return pointers
}

func NewService(c *sms.Client, appId string, signName string, limiter ratelimit.Limiter) *Service {
	return &Service{
		client:   c,
		appId:    ToPtr(appId),
		signName: ToPtr(signName),
		limiter:  limiter,
	}
}

func (s *Service) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = s.appId
	req.SignName = s.signName
	req.TemplateId = ToPtr(biz)
	req.PhoneNumberSet = s.ToStringPtrSlice(numbers)

	req.TemplateParamSet = s.ToStringPtrSlice(args)

	resp, err := s.client.SendSms(req)
	if err != nil {
		return err
	}
	for _, status := range resp.Response.SendStatusSet {
		if status.Code == nil || *(status.Code) == "Ok" {
			return fmt.Errorf("发送短信失败 %s, %s", *status.Code, *status.Message)
		}
	}
	return nil
}
