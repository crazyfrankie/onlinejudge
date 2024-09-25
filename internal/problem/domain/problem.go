package domain

type Problem struct {
	Id         uint64     `json:"id"`
	UserId     uint64     `json:"userId"`
	Title      string     `json:"title"`
	Tag        string     `json:"tag"`
	Content    string     `json:"content"`
	PassRate   string     `json:"passRate"`
	Prompt     []string   `json:"prompt"`
	TestCases  []TestCase `json:"testCases"`
	MaxMem     int        `json:"maxMem"`
	MaxRuntime int        `json:"maxRuntime"`
	Difficulty uint8      `json:"difficulty"`
}

type RoughProblem struct {
	Id       uint64 `json:"id"`
	Title    string `json:"title"`
	Tag      string `json:"tag"`
	PassRate string `json:"passRate"`
}

type TagWithCount struct {
	TagID        uint64 `json:"tag_id"`
	TagName      string `json:"tag_name"`
	ProblemCount int    `json:"problem_count"`
}

type Tag struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

type TestCase struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}
