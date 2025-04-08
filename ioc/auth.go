package ioc

import (
	"gorm.io/gorm"

	"github.com/crazyfrankie/framework-plugin/rbac"

	"github.com/crazyfrankie/onlinejudge/internal/middleware/auth"
)

func InitAuthz(db *gorm.DB) auth.Authorizer {
	a, err := rbac.NewAuthz(db)
	if err != nil {
		return nil
	}

	return a
}
