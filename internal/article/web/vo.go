package web

import "oj/internal/article/domain"

// View Object

type ListReq struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type ArticleReq struct {
	ID      uint64 `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (req ArticleReq) toDomain(uid uint64) domain.Article {
	return domain.Article{
		ID: req.ID,
		Author: domain.Author{
			Id: uid,
		},
		Title:   req.Title,
		Content: req.Content,
	}
}

type ListResp struct {
	ID         uint64
	Title      string
	Abstract   string
	AuthorID   uint64
	AuthorName string
	Status     uint8
	Ctime      string
	Utime      string
}

type DetailResp struct {
	ID         uint64
	Title      string
	Content    string
	AuthorID   uint64
	AuthorName string
	Status     uint8
	Ctime      string
	Utime      string
}
