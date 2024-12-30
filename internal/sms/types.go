package sms

import "github.com/crazyfrankie/onlinejudge/internal/sms/service"

type Service = service.Service
type Module struct {
	SmsSvc Service
}
