package ioc

import "oj/internal/user/service/oauth/github"

func InitGithubService() github.Service {
	return github.NewService()
}
