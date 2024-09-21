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
	"os"
	"os/exec"
	"path/filepath"

	"oj/internal/judgement/domain"
	"oj/internal/judgement/repository"
)

var (
	ErrSyntax = errors.New("your code not fit format")
)

type SubmitService interface {
	SubmitCode(ctx context.Context, submission domain.Submission, language string, inputs, outputs []string) (string, error)
	SetSubmission(id int8, code, stdin, expectCode string) Submission

	CodeFormat(language string, submission domain.Submission) (error, bool)
	GoFormat(code string, userId uint64) bool
	JavaFormat(code string, userId uint64) bool
	CppFormat(code string, userId uint64) bool
	PythonFormat(code string, userId uint64) bool

	GetFileWithUser(userId uint64, suffix string) string
}

type SubmissionSvc struct {
	repo     repository.SubmitRepository
	language map[string]int8
}

func NewSubmitService(repo repository.SubmitRepository) SubmitService {
	lang := map[string]int8{
		"Go":     95,
		"Java":   26,
		"Python": 71,
		"C++":    10,
	}
	return &SubmissionSvc{
		repo:     repo,
		language: lang,
	}
}

type Submission struct {
	LanguageId     int8   `json:"language_id"`
	Code           string `json:"source_code"`
	Stdin          string `json:"stdin"`           // 接收多个标准输入
	ExpectedOutput string `json:"expected_output"` // 接收多个期望输出
}

func (svc *SubmissionSvc) SubmitCode(ctx context.Context, submission domain.Submission, language string, inputs, outputs []string) (string, error) {
	// 判断语言类型
	id, ok := svc.language[language]
	if !ok {
		return "", errors.New("no language to fit")
	}

	// 代码格式检查
	err, done := svc.CodeFormat(language, submission)
	if !done {
		return "", err
	}

	// base64 编码
	encodedCode := base64.StdEncoding.EncodeToString([]byte(submission.Code))

	// 逐个提交测试用例
	for i, input := range inputs {
		expectedOutput := outputs[i]

		// 设置提交代码以及单个测试用例
		submit := svc.SetSubmission(id, encodedCode, input, expectedOutput)

		jsonData, err := json.Marshal(submit)
		if err != nil {
			return "", err
		}

		// 调用Judge0 API
		url := "https://judge0-ce.p.rapidapi.com/submissions?base64_encoded=true&wait=false&fields=*"
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", err
		}

		// 设置请求头
		req.Header.Add("x-rapidapi-key", "25e41db1b0msh5c2309187a14807p13cc02jsn21fce91931ce")
		req.Header.Add("x-rapidapi-host", "judge0-ce.p.rapidapi.com")
		req.Header.Add("Content-Type", "application/json")

		// 发送请求
		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer res.Body.Close()

		// 读取返回结果
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return "", err
		}

		// 处理返回的结果，可以将每次的结果存储或返回
		fmt.Println("Test case result:", string(body))

		// 如果需要缓存或存储结果，可以在这里进行处理
		// err = svc.repo.StoreEvaluation(ctx, string(body))
		// if err != nil {
		//     return string(body), err
		// }
	}

	return "All test cases submitted", nil
}

func (svc *SubmissionSvc) SetSubmission(id int8, code, stdin, expectCode string) Submission {
	submit := Submission{
		LanguageId:     id,
		Code:           code,
		Stdin:          stdin,      // 标准输入
		ExpectedOutput: expectCode, // 期望输出
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
		return ErrSyntax, true
	}
	return nil, false
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
