package domain

type Difficulty string

const (
	Easy   Difficulty = "Easy"
	Medium Difficulty = "Medium"
	Hard   Difficulty = "Hard"
)

type Problem struct {
	Id         uint64     `json:"id"`
	UserId     uint64     `json:"userId"`
	Title      string     `json:"title"`
	Tag        string     `json:"tag"`
	Content    string     `json:"content"`
	Prompt     []string   `json:"prompt"`
	PassRate   string     `json:"passRate"`
	MaxMem     int        `json:"maxMem"`
	MaxRuntime int        `json:"maxRuntime"`
	Difficulty Difficulty `json:"difficulty"`
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
