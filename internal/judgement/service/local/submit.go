package local

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/crazyfrankie/onlinejudge/internal/judgement/domain"
	"github.com/crazyfrankie/onlinejudge/internal/judgement/repository"
	domain2 "github.com/crazyfrankie/onlinejudge/internal/problem/domain"
	repository2 "github.com/crazyfrankie/onlinejudge/internal/problem/repository"
)

var (
	ErrSyntax = errors.New("your code not fit format")
)

type LocSubmitService interface {
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
func (svc *LocSubmitSvc) RunCode(ctx context.Context, submission domain.Submission) ([]domain.Evaluation, error) {
	ts, tmpl, err := svc.pmRepo.FindAllById(ctx, submission.ProblemID)
	if err != nil {
		return nil, err
	}

	temp := os.TempDir()
	// 用户源代码文件
	user, err := os.CreateTemp(temp, "main_*.go")
	if err != nil {
		return nil, err
	}
	defer os.Remove(user.Name())
	// 用户输出文件
	output, err := os.CreateTemp(temp, "user_out_*.txt")
	if err != nil {
		return nil, err
	}
	defer os.Remove(output.Name())

	err = svc.parseTemplate(ts, tmpl, submission.Code, user.Name())
	if err != nil {
		return nil, err
	}

	// 动态引入 import
	err = fixImport(user.Name())
	if err != nil {
		return nil, err
	}

	// 编译
	cmd := exec.Command("go", "build", user.Name())
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	// 可执行文件名称
	name := user.Name()[:len(user.Name())-2] + ".exe"
	defer os.Remove(name)

	err = Run(name, output.Name())
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func Run(execPath, output string) error {
	cmd := exec.Command(execPath)

	out, err := os.OpenFile(output, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer out.Close()
	cmd.Stdout = out
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// parseTemplate
// 拿到测试用例
// 拿到用户代码
// 构建模板变量
// 解析模板
// 执行模板渲染
// 写入 Go 文件
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

// CheckResult 由前端发起调用，轮询评测结果，CheckResult 不断从数据库中查询数据，直到评测结果插入
func (svc *LocSubmitSvc) CheckResult(ctx context.Context) error {
	return nil
}
