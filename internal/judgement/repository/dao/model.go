package dao

type Submission struct {
	Id         uint64 `gorm:"primaryKey,autoIncrement"`
	ProblemID  uint64 `gorm:"unique;not null"`
	UserId     uint64 `gorm:"unique;not null"`
	Code       string
	Language   string
	SubmitTime int64
	Ctime      int64
	Uptime     int64
	Deltime    int64
}

type Evaluation struct {
	Id           int64  `gorm:"primaryKey,autoIncrement"`
	SubmissionId uint64 `gorm:"index:problem_submit"`
	ProblemId    uint64 `gorm:"index:problem_submit"`
	Lang         string
	CpuTimeUsed  int64
	RealTimeUsed int64
	MemoryUsed   int64
	StatusMsg    string
	State        int8
	Ctime        int64
	Utime        int64
}

type State int8

const (
	PENGIND = "PENDING"
	SUCCESS = "SUCCESS"
	FAILED  = "FAILEd"
)

func (s State) toUint8(state string) int8 {
	switch state {
	case PENGIND:
		return 0
	case SUCCESS:
		return 1
	case FAILED:
		return 2
	}

	return -1
}

func (s State) toString() string {
	switch s {
	case 0:
		return PENGIND
	case 1:
		return SUCCESS
	case 2:
		return FAILED
	}

	return "unknown state"
}
