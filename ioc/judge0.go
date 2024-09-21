package ioc

import (
	"oj/internal/judgement/repository"
	"oj/internal/judgement/service"
	"os"
)

func InitJudgeService(repo repository.SubmitRepository) service.SubmitService {
	key, ok := os.LookupEnv("RAPIDAPI_KEY")
	if !ok {
		panic("environment variable rapidapiKey not found")
	}

	return service.NewSubmitService(repo, key)
}
