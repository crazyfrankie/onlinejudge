package domain

type Article struct {
	Title   string
	Content string
	Author  Author
}

type Author struct {
	Id   uint64
	Name string
}
