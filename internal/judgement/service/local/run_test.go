package local

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"text/template"
)

func TestTemplate(t *testing.T) {
	// 假设题目定义
	questionTemplate := `package main

import (
	"fmt"
)

func main() {
	testCases := []struct {
		input  []interface{}
		expect interface{}
	}{
		{{range .TestCases}}
		{input: []interface{}{ {{range .Input}} {{.}}, {{end}} }, expect: {{.Expect}} },
		{{end}}
	}

	for _, tc := range testCases {
		result := {{.FunctionName}}({{range $index, $value := .ParamNames}}{{if $index}}, {{end}}tc.input[{{$index}}]{{end}})
		fmt.Println(result, tc.expect)
	}
}

{{.UserCode}}
`

	// 假设用户提交代码
	userCode := `func findMid(nums []int, tg string) int {
	target, _ := strconv.Atoi(tg)
    left, right := 0, len(nums)-1
    for left <= right {
        mid := (left + right) / 2
        if nums[mid] > target {
            right = mid - 1
        } else if nums[mid] < target {
            left = mid + 1
        } else {
            return mid
        }
    }
    return -1
}`

	// 假设的测试用例
	testCases := []TestCase{
		{Input: []string{"[]int{1,2,3,4,5}", "2"}, Expect: "1"},
		{Input: []string{"[]int{10,20,30,40}", "25"}, Expect: "-1"},
	}
	// map[string]
	// 需要填充的模板变量
	data := TemplateData{
		FunctionName: "findMid",
		ParamNames:   []string{"[]int", "string"},
		TestCases:    testCases,
		UserCode:     userCode,
	}

	// 解析模板
	tmpl, err := template.New("code").Parse(questionTemplate)
	if err != nil {
		panic(err)
	}

	// 执行模板渲染
	var output bytes.Buffer
	err = tmpl.Execute(&output, data)
	if err != nil {
		panic(err)
	}

	// 写入 Go 文件
	err = os.WriteFile("user.go", output.Bytes(), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("Generated user.go successfully")

	std, err := os.OpenFile("./test/std_out.txt", os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer std.Close()

	err = fixImport("user.go")
	if err != nil {
		panic(err)
	}

	//cmd := exec.Command("go", "run", "user.go")
	//cmd.Stdout = std
	//if err := cmd.Start(); err != nil {
	//	fmt.Println("Error starting cmd:", err)
	//	return
	//}
	//if err := cmd.Wait(); err != nil {
	//	fmt.Println("Error running cmd:", err)
	//	return
	//}
}
