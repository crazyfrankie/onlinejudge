package service

import (
	"context"
	"oj/internal/judgement/domain"
)

const DockerFileTemp = ``

type SubmitService interface {
	SubmitCode(ctx context.Context, submission domain.Submission) (domain.Evaluation, error)
}

type SubmissionSvc struct {
}

func (svc *SubmissionSvc) SubmitCode(ctx context.Context, submission domain.Submission) (domain.Evaluation, error) {

}
