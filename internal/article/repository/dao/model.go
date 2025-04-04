package dao

// Article 制作库
type Article struct {
	ID      uint64 `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title   string `gorm:"type:varchar(1024)" bson:"title,omitempty"`
	Content string `gorm:"type:BLOB" bson:"content,omitempty"`
	// 如何设计索引
	// 在帖子这里，是什么样的查询场景
	// 对于创作者来说，需要看草稿箱，看到自己所有的文章
	// SELECT * FROM article WHERE author_id = 123 ORDER BY 'CTIME' DESC
	// 产品经理说，要按照创建时间的倒序排序
	// 单独查询某一篇 SELECT * FROM article WHERE id = 1
	// - 在 AuthorID 和 CTIME 上加联合索引
	// - 在 AuthorID 上创建索引
	AuthorID uint64 `gorm:"index:aid_ctime" bson:"author_id,omitempty"`
	Status   uint8  `bson:"status,omitempty"`
	Ctime    int64  `gorm:"index:aid_ctime" bson:"ctime,omitempty"`
	Utime    int64  `bson:"utime,omitempty"`
}

// OnlineArticle 线上库
type OnlineArticle struct {
	ID       uint64 `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title    string `gorm:"type:varchar(1024)" bson:"title,omitempty"`
	Content  string `gorm:"type:BLOB" bson:"content,omitempty"`
	AuthorID uint64 `gorm:"index:aid_time" bson:"author_id,omitempty"`
	Status   uint8  `bson:"status,omitempty"`
	Ctime    int64  `gorm:"index:aid_time" bson:"ctime,omitempty"`
	Utime    int64  `gorm:"index:aid_time" bson:"utime,omitempty"`
}

type Interactive struct {
	ID uint64 `gorm:"primaryKey,autoIncrement"`
	// 业务标识符
	BizID   uint64 `gorm:"uniqueIndex:biz_id_type"`
	Biz     string `gorm:"uniqueIndex:biz_id_type;type:varchar(128)"`
	ReadCnt int64
	LikeCnt int64
	Ctime   int64
	Utime   int64
}

type UserLike struct {
	ID    uint64 `gorm:"primaryKey,autoIncrement"`
	UID   uint64 `gorm:"uniqueIndex:uid_biz_id_type"`
	BizID uint64 `gorm:"uniqueIndex:uid_biz_id_type"`
	Biz   string `gorm:"uniqueIndex:uid_biz_id_type;type:varchar(128)"`
	Ctime int64
	Utime int64
	// 软删除
	Status uint8
}
