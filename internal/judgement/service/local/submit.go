package local

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	domain2 "oj/internal/problem/domain"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"oj/internal/judgement/domain"
	"oj/internal/judgement/repository"
	repository2 "oj/internal/problem/repository"
)

var (
	ErrSyntax = errors.New("your code not fit format")
)

type LocSubmitService interface {
	RunCode(ctx context.Context, submission domain.Submission, language string) ([]domain.Evaluation, error)
	//SubmitCode(ctx context.Context, submission domain.Submission, language string) ([]domain.Evaluation, error)

	ValidateCode(language, way string, submission domain.Submission) (error, bool)
	GoFormat(code, way string, userId, problemId uint64) bool
	JavaFormat(code, way string, userId, problemId uint64) bool
	CppFormat(code, way string, userId, problemId uint64) bool
	PythonFormat(code, way string, userId, problemId uint64) bool
	GetFileWithUser(userId, problemId uint64, suffix, way string) string
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
	// 先去查缓存
	evals, err := svc.repo.AcquireEvaluationResult(ctx, submission.UserId, submission.ProblemID)
	if err == nil {
		return evals, nil
	}

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
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			log.Printf("failed to remove temp directory: %v", err)
		}
	}(tempDir)

	tempFilePath := filepath.Join(tempDir, "main."+language)
	if err := os.WriteFile(tempFilePath, []byte(submission.Code), 0644); err != nil {
		return nil, errors.New("failed to write code to temp file")
	}

	// 获取测试用例
	testCases, err := svc.pmRepo.FindAllById(ctx, submission.ProblemID)
	if err != nil {
		return evals, fmt.Errorf("failed to get test cases: %w", err)
	}

	imageName, err := svc.RunDocker(language, tempDir, err)
	if err != nil {
		return evals, err
	}

	evals, err = svc.GetResult(testCases, imageName, evals)
	if err != nil {
		return evals, err
	}

	// 记录评估结果
	cacheErr := svc.repo.StoreEvaluationResult(ctx, submission.UserId, submission.ProblemID, evals)
	if cacheErr != nil {
		return nil, errors.New("failed to store execution result in cache")
	}

	return evals, nil
}

func (svc *LocSubmitSvc) GetResult(testCases []domain2.TestCase, imageName string, evals []domain.Evaluation) ([]domain.Evaluation, error) {
	// 运行 Docker 容器并传递输入用例
	for _, testCase := range testCases {
		var inputBuffer bytes.Buffer
		inputBuffer.WriteString(testCase.Input) // 将测试用例的输入写入缓冲区

		runCmd := exec.Command("sudo", "docker", "run", "--rm", "-i", "--runtime=runsc", imageName)
		runCmd.Stdin = &inputBuffer // 使用缓冲区作为标准输入

		fmt.Println("Input to container:", testCase.Input)

		output, err := runCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Container run error output: %s\n", string(output)) // 打印错误输出
			return nil, fmt.Errorf("code execution error: %s", err.Error())
		}

		fmt.Printf("Raw output from container: %s\n", string(output))

		// 解析输出，获取实际结果
		outputLines := strings.Split(strings.TrimSpace(string(output)), "\n")
		actualOutput := outputLines[0] // 假设第一行是实际结果，后续可以是其他信息

		// 对比输出与测试用例并构造评估结果
		status := "failed"
		if strings.TrimSpace(actualOutput) == strings.TrimSpace(testCase.Output) {
			status = "passed"
		}

		eval := domain.Evaluation{
			Status:  status,
			RunTime: "N/A", // 这里可以根据实际需要设置运行时间
			RunMem:  0,     // 这里可以根据实际需要设置内存使用
		}
		evals = append(evals, eval)
	}
	return evals, nil
}

func (svc *LocSubmitSvc) RunDocker(language string, tempDir string, err error) (string, error) {
	// 创建 Docker 镜像
	imageName := fmt.Sprintf("code_run_%d", time.Now().UnixNano())
	dockerfileContent := fmt.Sprintf(`
	FROM golang:1.23.2
	RUN mkdir /app
	WORKDIR /app
	COPY main.%s .
	RUN go build -o onlinejudge main.%s
	CMD ["./onlinejudge"]
	`, language, language)

	dockerfilePath := filepath.Join(tempDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		return "", errors.New("failed to write Dockerfile")
	}

	// 构建 Docker 镜像
	buildCmd := exec.Command("sudo", "docker", "build", "--no-cache", "-t", imageName, tempDir)
	outPut, err := buildCmd.CombinedOutput() // 获取输出
	if err != nil {
		return "", fmt.Errorf("failed to build Docker image: %s: %s", err.Error(), string(outPut))
	}

	fmt.Printf("Raw output from container: %s\n", string(outPut))
	return imageName, nil
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
	dir := filepath.Join(currDir, "internal", "judgement", "service", "local", "temp", "go")

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 构建文件的完整路径
<<<<<<< HEAD
	filePath := filepath.Join(dir, svc.GetFileWithUser(userId, problemId, way, "go"))
=======
	filePath := filepath.Join(dir, svc.GetRunFileWithUser(userId, problemId, "go"))
>>>>>>> origin/dev
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
	dir := filepath.Join(currDir, "internal", "judgement", "temp", "service", "local", "temp", "java")

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 构建文件的完整路径
<<<<<<< HEAD
	filePath := filepath.Join(dir, svc.GetFileWithUser(userId, problemId, way, "java"))
=======
	filePath := filepath.Join(dir, svc.GetRunFileWithUser(userId, problemId, "java"))
>>>>>>> origin/dev
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
	dir := filepath.Join(currDir, "internal", "judgement", "temp", "service", "local", "temp", "cpp")

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 构建文件的完整路径
<<<<<<< HEAD
	filePath := filepath.Join(dir, svc.GetFileWithUser(userId, problemId, way, "cpp"))
=======
	filePath := filepath.Join(dir, svc.GetRunFileWithUser(userId, problemId, "cpp"))
>>>>>>> origin/dev
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
	dir := filepath.Join(currDir, "internal", "judgement", "temp", "service", "local", "temp", "python")

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 构建文件的完整路径
<<<<<<< HEAD
	filePath := filepath.Join(dir, svc.GetFileWithUser(userId, problemId, way, "py"))
=======
	filePath := filepath.Join(dir, svc.GetRunFileWithUser(userId, problemId, "py"))
>>>>>>> origin/dev
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

func (svc *LocSubmitSvc) GetFileWithUser(userId, problemId uint64, suffix, way string) string {
	return fmt.Sprintf("user%dproblem:%d:%s.%s", userId, problemId, way, suffix)
}
