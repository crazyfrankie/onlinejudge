package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"oj/internal/judgement/domain"
	"oj/internal/judgement/repository"
	repository2 "oj/internal/problem/repository"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	ErrSyntax = errors.New("your code not fit format")
)

type SubmitService interface {
	RunCode(ctx context.Context, submission domain.Submission, language string) ([]string, error)
	RunResult(ctx context.Context, tokens []string) ([]domain.Evaluation, error)
	SetSubmission(id int8, code, stdin, stdout string) Submission
	SubmitResult(ctx context.Context, tokens []string) ([]domain.Evaluation, error)

	CodeFormat(language string, submission domain.Submission) (error, bool)
	GoFormat(code string, userId uint64) bool
	JavaFormat(code string, userId uint64) bool
	CppFormat(code string, userId uint64) bool
	PythonFormat(code string, userId uint64) bool

	GetFileWithUser(userId uint64, suffix string) string
	GetEvaluation(eval map[string]interface{}) (domain.Evaluation, error)
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

func (svc *SubmissionSvc) RunCode(ctx context.Context, submission domain.Submission, language string) ([]string, error) {
	// 判断语言类型
	id, ok := svc.language[language]
	if !ok {
		return []string{}, errors.New("no language to fit")
	}

	// 代码格式检查
	err, done := svc.CodeFormat(language, submission)
	if !done {
		return []string{}, err
	}

	// base64 编码
	encodedCode := base64.StdEncoding.EncodeToString([]byte(submission.Code))

	// 获取测试用例
	testCases, err := svc.pmRepo.FindById(ctx, submission.ProblemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get test cases: %w", err)
	}

	// 返回结果
	var results []string

	client := &http.Client{}
	// 逐个提交测试用例
	for _, tc := range testCases {
		// 设置提交代码以及单个测试用例
		submit := svc.SetSubmission(id, encodedCode, base64.StdEncoding.EncodeToString([]byte(tc.Input)), base64.StdEncoding.EncodeToString([]byte(tc.Output)))

		jsonData, err := json.Marshal(submit)
		if err != nil {
			return results, fmt.Errorf("failed to marshal submission: %w", err)
		}

		// 调用Judge0 API
		url := "https://judge0-ce.p.rapidapi.com/submissions?base64_encoded=true&wait=false&fields=*"
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return results, err
		}

		// 设置请求头
		req.Header.Add("x-rapidapi-key", svc.rapidapiKey)
		req.Header.Add("x-rapidapi-host", "judge0-ce.p.rapidapi.com")
		req.Header.Add("Content-Type", "application/json")

		// 发送请求
		res, err := client.Do(req)
		if err != nil {
			return results, err
		}

		// 读取返回结果
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return results, err
		}
		res.Body.Close() // 将 Close 移到这里

		// 将每次的结果返回
		results = append(results, string(body))
	}

	return results, nil
}

func (svc *SubmissionSvc) RunResult(ctx context.Context, tokens []string) ([]domain.Evaluation, error) {
	var evals []domain.Evaluation

	client := &http.Client{}
	for _, token := range tokens {
		url := fmt.Sprintf("https://judge0-ce.p.rapidapi.com/submissions/%s?base64_encoded=true&fields=*", token)
		fmt.Println(url)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return evals, err
		}

		req.Header.Add("x-rapidapi-key", "25e41db1b0msh5c2309187a14807p13cc02jsn21fce91931ce")
		req.Header.Add("x-rapidapi-host", "judge0-ce.p.rapidapi.com")

		res, err := client.Do(req)
		if err != nil {
			return evals, err
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return evals, err
		}
		res.Body.Close() // 确保响应体在函数结束时关闭

		var eval map[string]interface{}
		err = json.Unmarshal(body, &eval)
		if err != nil {
			return evals, err
		}

		// 获取所需字段
		evaluation, err2 := svc.GetEvaluation(eval)
		if err2 != nil {
			return evals, err2
		}

		evals = append(evals, evaluation)
	}

	return evals, nil
}

func (svc *SubmissionSvc) SubmitCode(ctx context.Context, submission domain.Submission, language string) ([]string, error) {

}

func (svc *SubmissionSvc) SubmitResult(ctx context.Context, tokens []string) ([]domain.Evaluation, error) {

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

func (svc *SubmissionSvc) SetSubmission(id int8, code, stdin, stdout string) Submission {
	submit := Submission{
		LanguageId: id,
		Code:       code,
		Stdin:      stdin,  // 标准输入
		StdOut:     stdout, // 期望输出
	}
	return submit
}

func (svc *SubmissionSvc) CodeFormat(language string, submission domain.Submission) (error, bool) {
	// 代码格式检查
	var formatOk bool
	switch language {
	case "Go":
		formatOk = svc.GoFormat(submission.Code, submission.UserId)
	case "Java":
		formatOk = svc.JavaFormat(submission.Code, submission.UserId)
	case "Python":
		formatOk = svc.PythonFormat(submission.Code, submission.UserId)
	case "C++":
		formatOk = svc.CppFormat(submission.Code, submission.UserId)
	}
	if !formatOk {
		return ErrSyntax, false
	}
	return nil, true
}

func (svc *SubmissionSvc) GoFormat(code string, userId uint64) bool {
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
	filePath := filepath.Join(dir, svc.GetFileWithUser(userId, "go"))

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

func (svc *SubmissionSvc) JavaFormat(code string, userId uint64) bool {
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
	filePath := filepath.Join(dir, svc.GetFileWithUser(userId, "java"))

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

func (svc *SubmissionSvc) CppFormat(code string, userId uint64) bool {
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
	filePath := filepath.Join(dir, svc.GetFileWithUser(userId, "cpp"))

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

func (svc *SubmissionSvc) PythonFormat(code string, userId uint64) bool {
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
	filePath := filepath.Join(dir, svc.GetFileWithUser(userId, "py"))

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

func (svc *SubmissionSvc) GetFileWithUser(userId uint64, suffix string) string {
	return fmt.Sprintf("user%d.%s", userId, suffix)
}
