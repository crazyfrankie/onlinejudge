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
	Status  string `json:"status"`
	RunTime string `json:"runTime"`
	RunMem  int    `json:"runMem"`
}
