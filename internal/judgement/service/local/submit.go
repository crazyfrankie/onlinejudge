package local

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/crazyfrankie/onlinejudge/internal/judgement/repository/dao"
	"strconv"
	"strings"

	"github.com/crazyfrankie/onlinejudge/internal/judgement/domain"
	"github.com/crazyfrankie/onlinejudge/internal/judgement/repository"
	repository2 "github.com/crazyfrankie/onlinejudge/internal/problem/repository"

	"github.com/crazyfrankie/go-judge/pkg/rpc"
)

var (
	ErrSyntax = errors.New("your code not fit format")
)

type LocSubmitService interface {
	RunCode(ctx context.Context, submission domain.Submission) (uint64, error)
	CheckResult(ctx context.Context, submitId uint64) (domain.Evaluation, error)
}

type LocSubmitSvc struct {
	repo   repository.LocalSubmitRepo
	pmRepo repository2.ProblemRepository
	client rpc.JudgeServiceClient
}

func NewLocSubmitService(repo repository.LocalSubmitRepo, pmRepo repository2.ProblemRepository, client rpc.JudgeServiceClient) LocSubmitService {
	return &LocSubmitSvc{
		repo:   repo,
		pmRepo: pmRepo,
		client: client,
	}
}

func (l *LocSubmitSvc) RunCode(ctx context.Context, submission domain.Submission) (uint64, error) {
	pm, err := l.pmRepo.FindProblemByID(ctx, submission.ProblemID)
	if err != nil {
		return 0, err
	}

	submission.CodeHash = hashCode(submission.Code)
	var submitID uint64
	submitID, err = l.repo.CreateSubmit(ctx, submission)
	if err != nil {
		return 0, err
	}

	err = l.repo.CreateEvaluate(ctx, domain.Evaluation{
		SubmissionId: submitID,
		ProblemId:    submission.ProblemID,
		Lang:         submission.Language,
		State:        "PENDING",
	})
	if err != nil {
		return 0, err
	}

	var res *rpc.JudgeResponse
	res, err = l.client.Judge(ctx, &rpc.JudgeRequest{
		Language:       getLanguage(submission.Language),
		ProblemId:      int64(submission.ProblemID),
		Uid:            int64(submission.UserId),
		Code:           submission.Code,
		FullTemplate:   pm.FullTemplate,
		TypeDefinition: pm.TypeDefinition,
		Input:          pm.Input,
		Output:         pm.Output,
		MaxMem:         strconv.Itoa(pm.MaxMem),
		MaxTime:        strconv.Itoa(pm.MaxRuntime),
	})
	if err != nil {
		return 0, err
	}

	//评测结果存入数据库
	var state dao.State
	err = l.repo.UpdateResult(ctx, submission.ProblemID, submitID, map[string]any{
		"cpu_time_used":  res.GetResult().TimeUsed,
		"real_time_used": res.GetResult().TimeUsed,
		"memory_used":    res.GetResult().MemoryUsed,
		"status_msg":     res.GetResult().StatusMsg,
		"state":          state.ToUint8(dao.PENGIND),
	})
	if err != nil {
		return 0, err
	}

	return submitID, nil
}

func (l *LocSubmitSvc) CheckResult(ctx context.Context, submitId uint64) (domain.Evaluation, error) {
	res, err := l.repo.FindEvaluate(ctx, submitId)
	if err != nil {
		return domain.Evaluation{}, err
	}

	return res, nil
}

func getLanguage(s string) rpc.Language {
	var lang rpc.Language
	switch s {
	case "go":
		lang = rpc.Language_go
	case "java":
		lang = rpc.Language_java
	case "cpp":
		lang = rpc.Language_cpp
	case "python":
		lang = rpc.Language_python
	}
	return lang
}

func hashCode(code string) string {
	preprocessed := preprocessCode(code)
	sum := sha256.Sum256([]byte(preprocessed))
	return fmt.Sprintf("%x", sum)
}

func preprocessCode(code string) string {
	var buf strings.Builder
	inBlockComment := false

	for _, line := range strings.Split(code, "\n") {
		trimmed := strings.TrimSpace(line)

		// 空行直接跳过
		if trimmed == "" {
			continue
		}

		// 处理多行注释（/* ... */）
		if strings.Contains(trimmed, "/*") {
			inBlockComment = true
		}
		if inBlockComment {
			if strings.Contains(trimmed, "*/") {
				inBlockComment = false
			}
			continue
		}

		// 单行注释过滤（C/Java）
		if idx := strings.Index(trimmed, "//"); idx != -1 {
			trimmed = trimmed[:idx]
			trimmed = strings.TrimSpace(trimmed)
		}

		// Python 注释（#）
		if idx := strings.Index(trimmed, "#"); idx != -1 {
			trimmed = trimmed[:idx]
			trimmed = strings.TrimSpace(trimmed)
		}

		if trimmed != "" {
			buf.WriteString(trimmed)
			buf.WriteString("\n")
		}
	}
	return buf.String()
}
