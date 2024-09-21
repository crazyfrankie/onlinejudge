package dao

type Problem struct {
	ID         uint64 `gorm:"primaryKey,autoIncrement"`
	Title      string `gorm:"unique;not null"`
	Content    string
	Difficulty uint8
	UserId     uint64
	Prompt     string
	PassRate   string
	MaxMem     int
	MaxRuntime int
	Ctime      int64
	Uptime     int64
	Deltime    int64
}

type Tag struct {
	ID   uint64 `gorm:"primaryKey,autoIncrement"`
	Name string `gorm:"unique;not null"`
}

type ProblemTag struct {
	ProblemID uint64 `gorm:"primaryKey,autoIncrement:false;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	TagID     uint64 `gorm:"primaryKey,autoIncrement:false;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}