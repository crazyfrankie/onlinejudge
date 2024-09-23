package ioc

import (
	"oj/internal/judgement/repository"
	"oj/internal/judgement/service"
	repository2 "oj/internal/problem/repository"
	"os"
)

func InitJudgeService(repo repository.SubmitRepository, pmRepo repository2.ProblemRepository) service.SubmitService {
	key, ok := os.LookupEnv("RAPIDAPI_KEY")
	if !ok {
		panic("environment variable rapidapiKey not found")
	}

	return service.NewSubmitService(repo, pmRepo, key)
}
