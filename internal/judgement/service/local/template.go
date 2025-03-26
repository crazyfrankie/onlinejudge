package local

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type TestCase struct {
	Input  []string
	Expect string
}

type TemplateData struct {
	FunctionName string
	ParamNames   []string
	TestCases    []TestCase
	UserCode     string
}

// fixImport 动态修复 import
func fixImport(filePath string) error {
	// 运行 go vet 检查缺少的包
	cmd := exec.Command("go", "vet", filePath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return nil // 没有错误，说明没有缺失的包
	}

	// 解析 go vet 的输出，提取缺失的包
	missingImports := extractMissingPackages(stderr.String())

	if len(missingImports) == 0 {
		return nil
	}

	// 读取文件内容
	code, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// 查找 package main 和 import ( 的位置
	pkgIndex := strings.Index(string(code), "package main")
	if pkgIndex == -1 {
		return fmt.Errorf("could not find package main in file")
	}

	importStartIndex := strings.Index(string(code), "import (")
	if importStartIndex == -1 {
		return fmt.Errorf("could not find import block in file")
	}

	// 查找 import ) 的位置
	importEndIndex := strings.Index(string(code), ")")
	if importEndIndex == -1 {
		return fmt.Errorf("could not find closing parenthesis for imports")
	}

	// 准备插入新的包，注意格式
	insertImports := ""
	for _, imp := range missingImports {
		// 确保每个包都与 fmt 对齐，并且是 tab 缀进，且换行
		insertImports += fmt.Sprintf("\n\t\"%s\"", imp)
	}

	// 查找 fmt 后的位置来插入新的包
	insertPosition := strings.LastIndex(string(code[:importEndIndex]), "\"fmt\"") + len("\"fmt\"")

	// 拼接最终代码：将缺失的包插入到 fmt 后面
	finalCode := string(code[:importEndIndex])[:insertPosition] + insertImports + "\n" + string(code[importEndIndex:])

	// 将修改后的内容写回文件
	err = os.WriteFile(filePath, []byte(finalCode), 0666)
	if err != nil {
		return fmt.Errorf("failed to write modified file: %w", err)
	}

	return nil
}

// extractMissingPackages 解析 `go vet` 的输出，提取缺失的包
func extractMissingPackages(output string) []string {
	// 正则匹配类似于 "undefined: strings" 这样的错误信息
	re := regexp.MustCompile(`undefined: (\w+)`)
	matches := re.FindAllStringSubmatch(output, -1)

	var packages []string
	for _, match := range matches {
		if len(match) > 1 {
			packages = append(packages, match[1]) // 提取标识符
		}
	}

	// 通过映射找出对应的标准库包（这里只列出部分）
	pkgMap := map[string]string{
		"strings": "strings",
		"math":    "math",
		"sort":    "sort",
		"time":    "time",
		"bytes":   "bytes",
		"strconv": "strconv",
		// 可继续补充其他常见包
	}

	var missingImports []string
	for _, id := range packages {
		if pkg, ok := pkgMap[id]; ok {
			missingImports = append(missingImports, pkg)
		}
	}
	return missingImports
}
