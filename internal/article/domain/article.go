package domain

import "time"

type Article struct {
	ID      uint64
	Title   string
	Content string
	Ctime   time.Time
	Utime   time.Time
	Author  Author
	Status  ArticleStatus
}

type Interactive struct {
	LikeCnt int64 `json:"like_cnt"`
	ReadCnt int64 `json:"read_cnt"`
	Liked   bool  `json:"liked"`
}

func (a Article) Abstract() string {
	// 摘要我们取前几句
	// 不能直接[:1024]
	// 可能会把中文截断
	cs := []rune(a.Content)
	if len(cs) < 100 {
		return a.Content
	}

	return string(cs[:100])
}

type ArticleStatus uint8

const (
	ArticleStatusUnknown ArticleStatus = iota
	ArticleStatusUnPublished
	ArticleStatusPublished
	ArticleStatusPrivate
)

func (s ArticleStatus) ToUint8() uint8 {
	return uint8(s)
}

func (s ArticleStatus) NonPublished() bool {
	return s != ArticleStatusPublished
}

func (s ArticleStatus) String() string {
	switch s {
	case ArticleStatusPrivate:
		return "private"
	case ArticleStatusUnPublished:
		return "unpublished"
	case ArticleStatusPublished:
		return "published"
	default:
		return "unknown"
	}
}

type Author struct {
	Id   uint64
	Name string
}
