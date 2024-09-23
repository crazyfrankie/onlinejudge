package dao

type Submission struct {
	Id         uint64 `gorm:"primaryKey,autoIncrement"`
	ProblemID  uint64 `gorm:"unique;not null"`
	UserId     uint64 `gorm:"unique;not null"`
	Code       string
	Language   string
	SubmitTime uint64
	Ctime      int64
	Uptime     int64
	Deltime    int64
}
