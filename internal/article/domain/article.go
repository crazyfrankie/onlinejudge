package domain

type Article struct {
	ID      uint64
	Title   string
	Content string
	Author  Author
}

type Author struct {
	Id   uint64
	Name string
}
