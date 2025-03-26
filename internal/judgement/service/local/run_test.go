package local

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"text/template"
)

func TestRunCode(t *testing.T) {
	// 读测试文件
	test, err := os.Open("./test/test_in.txt")
	if err != nil {
		panic(err)
	}
	defer test.Close()

	std, err := os.OpenFile("./test/std_out.txt", os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer std.Close()

	userCode := "func findMid(nums []int, target string) int {\n    left, right := 0, len(nums)-1\n    for left <= right {\n        mid := (left + right) / 2\n        if nums[mid] > target {\n			right = mid - 1\n		} else if nums[mid] < target {\n        	left = mid + 1\n		} else {\n			return mid\n		}\n    }\n    return -1\n}"

	code :=
		"package main\n\nimport (\n    \"fmt\"\n)\n\nfunc main() {\n    nums := []int{%s}\n    fmt.Println(findMid(nums, %s))\n}\n\n%s"

	// 创建一个扫描器逐行读取文件
	scanner := bufio.NewScanner(test)
	// 遍历每一行

	for scanner.Scan() {
		line := scanner.Text()

		lineArray := strings.Fields(line) // 也可以改成其他分隔符

		realCode := fmt.Sprintf(code, lineArray[0], lineArray[1], userCode)

		user, err := os.Create("user.go")
		//user, err := os.CreateTemp(os.TempDir(), "main_*.go")
		//if err != nil {
		//	panic(err)
		//}
		if err != nil {
			panic(err)
		}
		os.WriteFile(user.Name(), []byte(realCode), 0666)

		cmd := exec.Command("go", "run", user.Name())
		cmd.Stdout = std
		// 启动命令
		if err := cmd.Start(); err != nil {
			fmt.Println("Error starting cmd:", err)
			return
		}
		if err := cmd.Wait(); err != nil {
			fmt.Println("Error running cmd:", err)
			return
		}
		os.Remove(user.Name())
	}

	// 检查扫描过程中是否有错误
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

}

func TestReadFile(t *testing.T) {
	// 打开文件
	file, err := os.Open("./test/test_in.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// 用来存储每行数据的切片
	var lines [][]string

	// 创建一个扫描器逐行读取文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// 获取当前行的字符串
		line := scanner.Text()
		// 将每行的字符串分割成数组（例如按空格分割）
		lineArray := strings.Fields(line) // 也可以改成其他分隔符
		// 将该行的数组加入到lines数组中
		lines = append(lines, lineArray)
	}

	// 检查扫描过程中是否有错误
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// 输出每行数据
	for _, lineArray := range lines {
		fmt.Println(lineArray)
	}
}

//type TestCase struct {
//	Input  []string
//	Expect string
//}
//
//type TemplateData struct {
//	FunctionName string
//	ParamNames   []string
//	TestCases    []TestCase
//	UserCode     string
//}

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
		result := {{.FunctionName}}({{range $index, $value := .ParamNames}}{{if $index}}, {{end}}tc.input[{{$index}}].({{$value}}){{end}})
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
