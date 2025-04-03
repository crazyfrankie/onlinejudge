package dao

type Problem struct {
	ID           uint64 `gorm:"primaryKey,autoIncrement"`
	Title        string `gorm:"unique;not null"`
	Content      string
	Difficulty   uint8
	UserId       uint64
	PassRate     string
	TestCases    string `gorm:"type:json" json:"test_case"`
	Params       string
	TemplateCode string
	PreDefine    string
	MaxMem       int
	MaxRuntime   int
	Ctime        int64
	Uptime       int64
	Deltime      int64
}

type Tag struct {
	ID   uint64 `gorm:"primaryKey,autoIncrement"`
	Name string `gorm:"unique;not null"`
}

type ProblemTag struct {
	ProblemID uint64 `gorm:"primaryKey,autoIncrement:false;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	TagID     uint64 `gorm:"primaryKey,autoIncrement:false;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type TestCase struct {
	// 输入参数的类型信息
	InputTypes []string `json:"input_types"`
	// 输入参数的序列化值
	InputValues []string `json:"input_values"`
	// 期望输出的类型
	OutputType string `json:"output_type"`
	// 期望输出的序列化值
	OutputValue string `json:"output_value"`
	// 预定义的辅助函数或结构体
	Predefined string `json:"predefined"`
}

const (
	QuestionTemplate = `package main

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
		result := %s({{range $index, $value := .ParamNames}}{{if $index}}, {{end}}tc.input[{{$index}}].({{$value}}){{end}})
		fmt.Println(result, tc.expect)
	}
}

{{.UserCode}}
`
)
