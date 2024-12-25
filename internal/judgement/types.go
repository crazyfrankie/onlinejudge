package judgement

import "github.com/crazyfrankie/onlinejudge/internal/judgement/web"

type LocHandler = web.LocalSubmitHandler
type RemHandler = web.SubmissionHandler

type Module struct {
	LocHdl *LocHandler
	RemHdl *RemHandler
}
