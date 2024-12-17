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
	ID         uint64 `json:"id"`
	Title      string `json:"title"`
	Abstract   string `json:"abstract"`
	AuthorID   uint64 `json:"author_id"`
	AuthorName string `json:"author_name"`
	Status     uint8  `json:"status"`
	Ctime      string `json:"ctime"`
	Utime      string `json:"utime"`
}

type DetailResp struct {
	ID         uint64 `json:"ID"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	AuthorID   uint64 `json:"author_id"`
	AuthorName string `json:"author_name"`
	Status     uint8  `json:"status"`
	Ctime      string `json:"ctime"`
	Utime      string `json:"utime"`
}
