package remote

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"

	"github.com/crazyfrankie/onlinejudge/internal/judgement/domain"
	"github.com/crazyfrankie/onlinejudge/internal/judgement/repository"
	domain2 "github.com/crazyfrankie/onlinejudge/internal/problem/domain"
	repository2 "github.com/crazyfrankie/onlinejudge/internal/problem/repository"
)

var (
	ErrSyntax = errors.New("your code not fit format")
)

type SubmitService interface {
	RunCode(ctx context.Context, submission domain.Submission, language string) ([]domain.Evaluation, error)
	SubmitCode(ctx context.Context, submission domain.Submission, language string) ([]domain.Evaluation, error)

	GetEvaluation(eval map[string]interface{}) (domain.Evaluation, error)

	GetResult(ctx context.Context, testCases []domain2.TestCase, langId int8, encodedCode string, result []domain.Evaluation) ([]domain.Evaluation, error)
	SetSubmission(id int8, code, stdin, stdout string) Submission

	CodeFormat(language, way string, submission domain.Submission) (error, bool)
	GoFormat(code, way string, userId, problemId uint64) bool
	JavaFormat(code, way string, userId, problemId uint64) bool
	CppFormat(code, way string, userId, problemId uint64) bool
	PythonFormat(code, way string, userId, problemId uint64) bool

	GetFileWithUser(userId, problemId uint64, suffix, way string) string
}

type SubmissionSvc struct {
	repo        repository.SubmitRepository
	pmRepo      repository2.ProblemRepository
	language    map[string]int8
	rapidapiKey string
}

func NewSubmitService(repo repository.SubmitRepository, pmRepo repository2.ProblemRepository, key string) SubmitService {
	lang := map[string]int8{
		"Go":     95,
		"Java":   26,
		"Python": 71,
		"C++":    10,
	}
	return &SubmissionSvc{
		repo:        repo,
		pmRepo:      pmRepo,
		language:    lang,
		rapidapiKey: key,
	}
}

type Submission struct {
	LanguageId int8   `json:"language_id"`
	Code       string `json:"source_code"`
	Stdin      string `json:"stdin"`  // 接收多个标准输入
	StdOut     string `json:"stdout"` // 接收多个期望输出
}

func (svc *SubmissionSvc) RunCode(ctx context.Context, submission domain.Submission, language string) ([]domain.Evaluation, error) {
	//先查缓存 如果有则直接返回
	hash := sha256.Sum256([]byte(submission.Code))
	hashString := hex.EncodeToString(hash[:])

	evals, err := svc.repo.AcquireEvaluation(ctx, submission.UserId, hashString)
	if err == nil {
		return evals, err
	}

	//判断语言类型
	id, ok := svc.language[language]
	if !ok {
		return evals, errors.New("no language to fit")
	}

	// 代码格式检查
	err, done := svc.CodeFormat(language, "run", submission)

	if !done {
		return evals, err
	}

	// base64 编码
	encodedCode := base64.StdEncoding.EncodeToString([]byte(submission.Code))

	// 获取测试用例
	testCases, err := svc.pmRepo.FindById(ctx, submission.ProblemID)
	if err != nil {
		return evals, fmt.Errorf("failed to get test cases: %w", err)
	}

	// 获取返回结果
	evals, err = svc.GetResult(ctx, testCases, id, encodedCode, evals)

	// 更新缓存
	var updating int32

	if atomic.LoadInt32(&updating) == 0 {
		if atomic.CompareAndSwapInt32(&updating, 0, 1) {
			go func() {
				// 更新逻辑
				svc.repo.StoreEvaluation(ctx, submission.UserId, hashString, evals)
				atomic.StoreInt32(&updating, 0) // 更新完成，重置状态
			}()
		}
	}

	return evals, err
}

func (svc *SubmissionSvc) SubmitCode(ctx context.Context, submission domain.Submission, language string) ([]domain.Evaluation, error) {
	var evals []domain.Evaluation

	//判断语言类型
	id, ok := svc.language[language]
	if !ok {
		return evals, errors.New("no language to fit")
	}

	// 代码格式检查
	err, done := svc.CodeFormat(language, "submit", submission)
	if !done {
		return evals, err
	}

	// base64 编码
	encodedCode := base64.StdEncoding.EncodeToString([]byte(submission.Code))

	// 获取测试用例
	testCases, _, err := svc.pmRepo.FindAllById(ctx, submission.ProblemID)
	if err != nil {
		return evals, fmt.Errorf("failed to get test cases: %w", err)
	}

	// 获取返回结果
	evals, err = svc.GetResult(ctx, testCases, id, encodedCode, evals)

	return evals, err
}

func (svc *SubmissionSvc) GetResult(ctx context.Context, testCases []domain2.TestCase, langId int8, encodedCode string, result []domain.Evaluation) ([]domain.Evaluation, error) {
	// 逐个提交测试用例
	for _, tc := range testCases {
		// 设置提交代码以及单个测试用例
		submit := svc.SetSubmission(langId, encodedCode, base64.StdEncoding.EncodeToString([]byte(tc.Input)), base64.StdEncoding.EncodeToString([]byte(tc.Output)))

		jsonData, err := json.Marshal(submit)
		if err != nil {
			return result, fmt.Errorf("failed to marshal submission: %w", err)
		}

		// 调用Judge0 API
		url := "https://judge0-ce.p.rapidapi.com/submissions?base64_encoded=true&wait=true&fields=*"
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
		if err != nil {
			return result, err
		}

		// 设置请求头
		req.Header.Add("x-rapidapi-key", svc.rapidapiKey)
		req.Header.Add("x-rapidapi-host", "judge0-ce.p.rapidapi.com")
		req.Header.Add("Content-Type", "application/json")

		// 发送请求
		client := http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return result, err
		}

		// 读取返回结果
		var eval map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&eval); err != nil {
			return result, fmt.Errorf("failed to decode response: %w", err)
		}

		res.Body.Close() // 确保响应体在函数结束时关闭

		evaluation, err := svc.GetEvaluation(eval)
		result = append(result, evaluation)
	}

	return result, nil
}

func (svc *SubmissionSvc) SetSubmission(id int8, code, stdin, stdout string) Submission {
	submit := Submission{
		LanguageId: id,
		Code:       code,
		Stdin:      stdin,  // 标准输入
		StdOut:     stdout, // 期望输出
	}
	return submit
}

func (svc *SubmissionSvc) CodeFormat(language, way string, submission domain.Submission) (error, bool) {
	// 代码格式检查
	var formatOk bool
	switch language {
	case "Go":
		formatOk = svc.GoFormat(submission.Code, way, submission.ProblemID, submission.UserId)
	case "Java":
		formatOk = svc.JavaFormat(submission.Code, way, submission.ProblemID, submission.UserId)
	case "Python":
		formatOk = svc.PythonFormat(submission.Code, way, submission.ProblemID, submission.UserId)
	case "C++":
		formatOk = svc.CppFormat(submission.Code, way, submission.ProblemID, submission.UserId)
	}
	if !formatOk {
		return ErrSyntax, false
	}
	return nil, true
}

func (svc *SubmissionSvc) GoFormat(code, way string, userId, problemId uint64) bool {
	// 定义目标文件夹路径
	currDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 构建路径
	dir := filepath.Join(currDir, "internal", "judgement", "service", "remote", "temp", "go")

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 构建文件的完整路径
	filePath := filepath.Join(dir, svc.GetFileWithUser(userId, problemId, way, "go"))
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

func (svc *SubmissionSvc) JavaFormat(code, way string, userId, problemId uint64) bool {
	// 定义目标文件夹路径
	currDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 构建路径
	dir := filepath.Join(currDir, "internal", "judgement", "service", "remote", "temp", "java")

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 构建文件的完整路径
	filePath := filepath.Join(dir, svc.GetFileWithUser(userId, problemId, way, "java"))
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

func (svc *SubmissionSvc) CppFormat(code, way string, userId, problemId uint64) bool {
	// 定义目标文件夹路径
	currDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 构建路径
	dir := filepath.Join(currDir, "internal", "judgement", "service", "remote", "temp", "cpp")

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 构建文件的完整路径
	filePath := filepath.Join(dir, svc.GetFileWithUser(userId, problemId, way, "cpp"))
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

func (svc *SubmissionSvc) PythonFormat(code, way string, userId, problemId uint64) bool {
	// 定义目标文件夹路径
	currDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 构建路径
	dir := filepath.Join(currDir, "internal", "judgement", "service", "remote", "temp", "python")

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 构建文件的完整路径
	filePath := filepath.Join(dir, svc.GetFileWithUser(userId, problemId, way, "py"))
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

func (svc *SubmissionSvc) GetFileWithUser(userId, problemId uint64, suffix, way string) string {
	return fmt.Sprintf("user%dproblem:%d:%s.%s", userId, problemId, way, suffix)
}

func (svc *SubmissionSvc) GetEvaluation(eval map[string]interface{}) (domain.Evaluation, error) {
	status, ok := eval["status"].(map[string]interface{})
	if !ok {
		return domain.Evaluation{}, errors.New("invalid status format")
	}
	description, ok := status["description"].(string)
	if !ok {
		return domain.Evaluation{}, errors.New("invalid description format")
	}
	runMem, ok := eval["memory"].(float64)
	if !ok {
		return domain.Evaluation{}, errors.New("invalid memory format")
	}
	runTime, ok := eval["time"].(string)
	if !ok {
		return domain.Evaluation{}, errors.New("invalid time format")
	}

	evaluation := domain.Evaluation{
		Status:  description,
		RunMem:  int(runMem),
		RunTime: runTime,
	}
	return evaluation, nil
}
