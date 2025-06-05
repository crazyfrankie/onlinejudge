package dao

// ChunkSize 定义分片大小为1MB
const ChunkSize = 1 * 1024 * 1024 // 1MB

// ArticleChunk 文章分片
type ArticleChunk struct {
	ID        int64  `bson:"id,omitempty"`         // 分片ID
	ArticleID uint64 `bson:"article_id,omitempty"` // 文章ID
	Content   string `bson:"content,omitempty"`    // 分片内容
	Order     int    `bson:"order,omitempty"`      // 分片顺序
	Ctime     int64  `bson:"ctime,omitempty"`
	Utime     int64  `bson:"utime,omitempty"`
}

type MongoArticle struct {
	ID         uint64 `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title      string `gorm:"type:varchar(1024)" bson:"title,omitempty"`
	Content    string `gorm:"type:BLOB" bson:"content,omitempty"`
	AuthorID   uint64 `gorm:"index:aid_ctime" bson:"author_id,omitempty"`
	Status     uint8  `bson:"status,omitempty"`
	Abstract   string `bson:"abstract,omitempty"`
	Ctime      int64  `gorm:"index:aid_ctime" bson:"ctime,omitempty"`
	Utime      int64  `bson:"utime,omitempty"`
	ChunkCount int    `bson:"chunk_count,omitempty"` // 分片数量，0表示未分片
}

// Article 制作库
type Article struct {
	ID       uint64 `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title    string `gorm:"type:varchar(1024)" bson:"title,omitempty"`
	Content  string `gorm:"type:BLOB" bson:"content,omitempty"`
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
