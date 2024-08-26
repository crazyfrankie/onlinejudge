package domain

type Difficulty string

const (
	Easy   Difficulty = "Easy"
	Medium Difficulty = "Medium"
	Hard   Difficulty = "Hard"
)

type Problem struct {
	Id         uint64
	Title      string
	Content    string
	Prompt     []string
	PassRate   string
	Difficulty Difficulty
}
