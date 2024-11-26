package domain

type Article struct {
	ID      uint64
	Title   string
	Content string
	Author  Author
	Status  ArticleStatus
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
