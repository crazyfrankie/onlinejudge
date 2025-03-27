package local

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/crazyfrankie/onlinejudge/internal/judgement/domain"
	"github.com/crazyfrankie/onlinejudge/internal/judgement/repository"
	domain2 "github.com/crazyfrankie/onlinejudge/internal/problem/domain"
	repository2 "github.com/crazyfrankie/onlinejudge/internal/problem/repository"

	"github.com/crazyfrankie/judge-go"
)

var (
	ErrSyntax = errors.New("your code not fit format")
)

type LocSubmitService interface {
	RunCode(ctx context.Context, submission *domain.Submission) (uint64, error)
	CheckResult(ctx context.Context, submitId uint64) (domain.Evaluation, error)
}

type LocSubmitSvc struct {
	repo   repository.LocalSubmitRepo
	pmRepo repository2.ProblemRepository
}

func NewLocSubmitService(repo repository.LocalSubmitRepo, pmRepo repository2.ProblemRepository) LocSubmitService {
	return &LocSubmitSvc{
		repo:   repo,
		pmRepo: pmRepo,
	}
}

//RunCode 运行前端提交的代码，并写入结果到数据库
func (svc *LocSubmitSvc) RunCode(ctx context.Context, submission *domain.Submission) (uint64, error) {
	ts, tmpl, err := svc.pmRepo.FindAllById(ctx, submission.ProblemID)
	if err != nil {
		return 0, err
	}

	temp := os.TempDir()
	// 用户源代码文件
	user, err := os.CreateTemp(temp, "main_*.go")
	if err != nil {
		return 0, err
	}
	defer os.Remove(user.Name())
	// 用户输出文件
	output, err := os.CreateTemp(temp, "user_out_*.txt")
	if err != nil {
		return 0, err
	}
	defer os.Remove(output.Name())

	err = svc.parseTemplate(ts, tmpl, submission.Code, user.Name())
	if err != nil {
		return 0, err
	}

	// 动态引入 import
	err = fixImport(user.Name())
	if err != nil {
		return 0, err
	}

	// 编译
	cmd := exec.Command("go", "build", user.Name())
	if err := cmd.Run(); err != nil {
		return 0, err
	}
	// 可执行文件名称
	name := user.Name()[:len(user.Name())-2] + ".exe"
	defer os.Remove(name)

	// 创建评测实例
	jd := createJudge(name, output.Name())
	_, err = jd.Run(ctx)
	if err != nil {
		return 0, err
	}

	// 评测结果存入数据库
	// TODO

	_, err = jd.Check()
	if err != nil {
		return 0, err
	}

	// 修改数据库状态
	// TODO

	return 0, nil
}

func createJudge(execPath, userOut string) *judge.Judge {
	jg := judge.NewJudge(&judge.Config{
		Limits: struct {
			CPU    time.Duration
			Memory int64
			Stack  int64
			Output int64
		}{
			CPU:    2 * time.Second,
			Memory: 128 * 1024 * 1024,
			Stack:  8 * 1024 * 1024,
			Output: 10 * 1024 * 1024,
		},
		Exec: struct {
			Path string
			Args []string
			Env  []string
		}{
			Path: "./" + execPath,
		},
		Files: struct {
			UserOutput string
			CgroupPath string
		}{
			UserOutput: userOut,
			CgroupPath: "cgroup",
		},
	})

	return jg
}

// parseTemplate
// Get the test case
// Get the user code
// Build the template variables
// Parsing the template
// Perform template rendering
// Write the Go file
func (svc *LocSubmitSvc) parseTemplate(testCases []domain2.TestCase, tmplCode, userCode, targetFile string) error {
	// 构建测试用例
	tcs := make([]TestCase, len(testCases))
	for _, tc := range testCases {
		inputs := strings.Fields(tc.Input)
		tc := TestCase{
			Input:  inputs,
			Expect: tc.Output,
		}

		tcs = append(tcs, tc)
	}

	// 构建模板变量
	data := TemplateData{
		ParamNames: []string{"[]int", "int"},
		TestCases:  tcs,
		UserCode:   userCode,
	}

	// 解析
	tmpl, err := template.New("code").Parse(tmplCode)
	if err != nil {
		return err
	}

	// 渲染
	var output bytes.Buffer
	err = tmpl.Execute(&output, data)
	if err != nil {
		return err
	}

	// 写入
	err = os.WriteFile(targetFile, output.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}

// CheckResult The front-end initiates the call, polling for evaluation results,
// and CheckResult keeps querying the data from
// the database until the evaluation results are inserted.
func (svc *LocSubmitSvc) CheckResult(ctx context.Context, submitId uint64) (domain.Evaluation, error) {
	// 查询数据库
	return domain.Evaluation{}, nil
}
