package local

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	repository2 "oj/internal/problem/repository"
	"os"
	"os/exec"
	"path/filepath"

	"oj/internal/judgement/domain"
	"oj/internal/judgement/repository"
)

var (
	ErrSyntax = errors.New("your code not fit format")
)

type LocSubmitService interface {
	RunCode(ctx context.Context, submission domain.Submission, language string) ([]domain.Evaluation, error)
	//SubmitCode(ctx context.Context, submission domain.Submission, language string) ([]domain.Evaluation, error)
}

type LocSubmitSvc struct {
	repo     repository.LocalSubmitRepo
	pmRepo   repository2.ProblemRepository
	language map[string]int8
}

func NewLocSubmitService(repo repository.LocalSubmitRepo, pmRepo repository2.ProblemRepository) LocSubmitService {
	lang := map[string]int8{
		"cpp":  1,
		"go":   2,
		"java": 3,
		"py":   4,
	}
	return &LocSubmitSvc{
		repo:     repo,
		pmRepo:   pmRepo,
		language: lang,
	}
}

func (svc *LocSubmitSvc) RunCode(ctx context.Context, submission domain.Submission, language string) ([]domain.Evaluation, error) {
	var evals []domain.Evaluation

	// 语言支持验证
	_, ok := svc.language[language]
	if !ok {
		return evals, errors.New("no language to fit")
	}

	// 代码格式检查
	err, done := svc.ValidateCode(language, "run", submission)
	if !done {
		return evals, err
	}

	// 创建临时文件保存代码
	tempDir, err := os.MkdirTemp("", "code_run_*")
	if err != nil {
		return nil, errors.New("failed to create temp directory")
	}
	defer os.RemoveAll(tempDir)

	tempFilePath := filepath.Join(tempDir, "main."+language)
	if err := os.WriteFile(tempFilePath, []byte(submission.Code), 0644); err != nil {
		return nil, errors.New("failed to write code to temp file")
	}

	// 获取测试用例
	testCases, err := svc.pmRepo.FindById(ctx, submission.ProblemID)
	if err != nil {
		return evals, fmt.Errorf("failed to get test cases: %w", err)
	}

	// 执行代码并获取输出
	sandboxCmd := exec.Command("/path/to/sandbox/executable", tempFilePath)
	output, err := sandboxCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("code execution error: %s", err.Error())
	}

	// 假设输出是 JSON 格式
	var result struct {
		Stdout  string `json:"stdout"`
		RunTime string `json:"runtime"`
		RunMem  int    `json:"memory"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, errors.New("failed to parse sandbox output")
	}

	// 对比输出与测试用例并构造评估结果
	for _, testCase := range testCases {
		expectedOutput := testCase.Output
		actualOutput := result.Stdout // 从沙箱结果中获取实际输出

		status := "failed"
		if expectedOutput == actualOutput {
			status = "passed"
		}

		eval := domain.Evaluation{
			Status:  status,
			RunTime: result.RunTime,
			RunMem:  result.RunMem,
		}
		evals = append(evals, eval)
	}

	// 记录评估结果
	cacheErr := svc.repo.StoreEvaluationResult(ctx, submission.UserId, submission.ProblemID, evals)
	if cacheErr != nil {
		return nil, errors.New("failed to store execution result in cache")
	}

	return evals, nil
}

//func (svc *LocSubmitSvc) SubmitCode(ctx context.Context, submission domain.Submission, language string) ([]domain.Evaluation, error) {
//
//}

func (svc *LocSubmitSvc) ValidateCode(language, way string, submission domain.Submission) (error, bool) {
	// 代码格式检查
	var formatOk bool
	switch language {
	case "go":
		formatOk = svc.GoFormat(submission.Code, way, submission.ProblemID, submission.UserId)
	case "java":
		formatOk = svc.JavaFormat(submission.Code, way, submission.ProblemID, submission.UserId)
	case "py":
		formatOk = svc.PythonFormat(submission.Code, way, submission.ProblemID, submission.UserId)
	case "cpp":
		formatOk = svc.CppFormat(submission.Code, way, submission.ProblemID, submission.UserId)
	}
	if !formatOk {
		return ErrSyntax, false
	}
	return nil, true
}

func (svc *LocSubmitSvc) GoFormat(code, way string, userId, problemId uint64) bool {
	// 定义目标文件夹路径
	currDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 构建路径
	dir := filepath.Join(currDir, "internal", "judgement", "service", "temp", "go")

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 构建 temp.go 文件的完整路径
	// 构建文件的完整路径
	var filePath string
	if way == "run" {
		filePath = filepath.Join(dir, svc.GetRunFileWithUser(userId, problemId, "go"))
	} else {
		filePath = filepath.Join(dir, svc.GetRunFileWithUser(userId, problemId, "go"))
	}

	// 将代码写入 temp.go 文件中
	err = os.WriteFile(filePath, []byte(code), 0644)
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 使用 go vet 或 go build 来检测语法
	cmd := exec.Command("go", "vet", filePath)
	_, err = cmd.CombinedOutput()
	if err != nil {
		return false
	}

	return true
}

func (svc *LocSubmitSvc) JavaFormat(code, way string, userId, problemId uint64) bool {
	// 定义目标文件夹路径
	currDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 构建路径
	dir := filepath.Join(currDir, "internal", "judgement", "service", "temp", "java")

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 构建文件的完整路径
	var filePath string
	if way == "run" {
		filePath = filepath.Join(dir, svc.GetRunFileWithUser(userId, problemId, "java"))
	} else {
		filePath = filepath.Join(dir, svc.GetRunFileWithUser(userId, problemId, "java"))
	}

	// 将代码写入文件中
	err = os.WriteFile(filePath, []byte(code), 0644)
	if err != nil {
		log.Fatal(err)
		return false
	}

	cmd := exec.Command("javac", filePath)
	_, err = cmd.CombinedOutput()
	if err != nil {
		return false
	}

	return true
}

func (svc *LocSubmitSvc) CppFormat(code, way string, userId, problemId uint64) bool {
	// 定义目标文件夹路径
	currDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 构建路径
	dir := filepath.Join(currDir, "internal", "judgement", "service", "temp", "cpp")

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 构建文件的完整路径
	var filePath string
	if way == "run" {
		filePath = filepath.Join(dir, svc.GetRunFileWithUser(userId, problemId, "cpp"))
	} else {
		filePath = filepath.Join(dir, svc.GetRunFileWithUser(userId, problemId, "cpp"))
	}

	// 将代码写入文件中
	err = os.WriteFile(filePath, []byte(code), 0644)
	if err != nil {
		log.Fatal(err)
		return false
	}

	cmd := exec.Command("g++", "-fsyntax-only", filePath)
	_, err = cmd.CombinedOutput()
	if err != nil {
		return false
	}

	return true
}

func (svc *LocSubmitSvc) PythonFormat(code, way string, userId, problemId uint64) bool {
	// 定义目标文件夹路径
	currDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 构建路径
	dir := filepath.Join(currDir, "internal", "judgement", "service", "temp", "python")

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 构建文件的完整路径
	var filePath string
	if way == "run" {
		filePath = filepath.Join(dir, svc.GetRunFileWithUser(userId, problemId, "py"))
	} else {
		filePath = filepath.Join(dir, svc.GetRunFileWithUser(userId, problemId, "py"))
	}

	// 将代码写入文件中
	err = os.WriteFile(filePath, []byte(code), 0644)
	if err != nil {
		log.Fatal(err)
		return false
	}

	cmd := exec.Command("python3", "-m", "py_compile", filePath)
	_, err = cmd.CombinedOutput()
	if err != nil {
		return false
	}

	return true
}

func (svc *LocSubmitSvc) GetRunFileWithUser(userId, problemId uint64, suffix string) string {
	return fmt.Sprintf("user%dproblem:%d:run.%s", userId, problemId, suffix)
}
