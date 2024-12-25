package problem

import (
	"github.com/crazyfrankie/onlinejudge/internal/problem/repository"
	"github.com/crazyfrankie/onlinejudge/internal/problem/web"
)

type Handler = web.ProblemHandler
type Repository = repository.ProblemRepository

type Module struct {
	Hdl  *Handler
	Repo Repository
}
