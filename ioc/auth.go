package ioc

import (
	"gorm.io/gorm"
	"time"

	"github.com/crazyfrankie/framework-plugin/rbac"

	"github.com/crazyfrankie/onlinejudge/internal/middleware/auth"
)

func InitAuthz(db *gorm.DB) auth.Authorizer {
	a, err := rbac.NewAuthz(db, rbac.WithLoadTime(time.Second*30))
	if err != nil {
		return nil
	}

	return a
}
