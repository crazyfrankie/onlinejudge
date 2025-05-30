package dao

type Problem struct {
	ID             uint64 `gorm:"primaryKey,autoIncrement"`
	Title          string `gorm:"type:varchar(128)"`
	Content        string
	FullTemplate   string
	TypeDefinition string
	Func           string
	Inputs         string
	Outputs        string
	Difficulty     string `gorm:"type:varchar(20)"`
	TotalSubmit    int64  `gorm:"not null,default:0"`
	TotalPass      int64  `gorm:"not null,default:0"`
	MaxMem         int    `gorm:"not null,default:0"`
	MaxRuntime     int    `gorm:"not null,default:0"`
	Ctime          int64
	Utime          int64
}

type Tag struct {
	ID   uint64 `gorm:"primaryKey,autoIncrement"`
	Name string `gorm:"unique;not null"`
}

type ProblemTag struct {
	ProblemID uint64 `gorm:"primaryKey,autoIncrement:false;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	TagID     uint64 `gorm:"primaryKey,autoIncrement:false;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
