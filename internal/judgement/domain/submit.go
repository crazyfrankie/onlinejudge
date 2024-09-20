package domain

type Submission struct {
	Id         uint64 `json:"id"`
	ProblemID  uint64 `json:"problemID"`
	UserId     uint64 `json:"userId"`
	Code       string `json:"code"`
	Language   string `json:"language"`
	SubmitTime uint64 `json:"submitTime"`
}

type Evaluation struct {
	SubmissionId uint64 `json:"submissionId"`
	Status       string `json:"status"`
	RunTime      uint64 `json:"runTime"`
	RunMem       uint64 `json:"runMem"`
	TestFields   []bool `json:"testFields"`
}
