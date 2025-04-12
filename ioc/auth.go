package ioc

import (
	"github.com/crazyfrankie/onlinejudge/internal/mws"
	"gorm.io/gorm"
	"time"

	"github.com/crazyfrankie/framework-plugin/rbac"
)

func InitAuthz(db *gorm.DB) mws.Authorizer {
	a, err := rbac.NewAuthz(db, rbac.WithLoadTime(time.Second*60))
	if err != nil {
		return nil
	}

	return a
}
