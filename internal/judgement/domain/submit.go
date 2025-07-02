package domain

type Submission struct {
	Id         uint64 `json:"id"`
	ProblemID  uint64 `json:"problemID"`
	UserId     uint64 `json:"userId"`
	Code       string `json:"code"`
	CodeHash   string `json:"codeHash"`
	Language   string `json:"language"`
	SubmitTime int64  `json:"submitTime"`
}

type Evaluation struct {
	Id           int64  `json:"id"`
	SubmissionId uint64 `json:"submission_id"`
	ProblemId    uint64 `json:"problem_id"`
	Lang         string `json:"lang"`
	CpuTimeUsed  int64  `json:"cpu_time_used"`
	RealTimeUsed int64  `json:"real_time_used"`
	MemoryUsed   int64  `json:"memory_used"`
	StatusMsg    string `json:"status_msg"`
	State        string `json:"state"`
}

type RemoteEvaluation struct {
	RunMem  int64  `json:"run_mem"`
	RunTime string `json:"run_time"`
	Msg     string `json:"msg"`
}
