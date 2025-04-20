package sm

import "github.com/crazyfrankie/onlinejudge/internal/sm/service"

type SmSvc = service.CodeService

type Module struct {
	Sm SmSvc
}
