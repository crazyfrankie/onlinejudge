package constant

type ErrorCode struct {
	Code    int32
	Message string
}

// 通用错误
var (
	Success = ErrorCode{Code: 00000, Message: "success"}

	ErrInvalidParams  = ErrorCode{Code: 00400, Message: "invalid parameters"}
	ErrInternalServer = ErrorCode{Code: 00500, Message: "internal server errors"}
)

// 身份校验错误
var (
	ErrInvalidToken = ErrorCode{Code: 10400, Message: "invalid token"}
	ErrTokenExpired = ErrorCode{Code: 10401, Message: "token expired"}
	ErrLoginYet     = ErrorCode{Code: 10402, Message: "have not logged in yet"}
	ErrSessExpired  = ErrorCode{Code: 10403, Message: "session expired"}
	ErrUnauthorized = ErrorCode{Code: 10404, Message: "unauthorized"}
	ErrForbidden    = ErrorCode{Code: 10405, Message: "forbidden"}
)

// 用户相关错误
var (
	ErrInvalidCredentials = ErrorCode{Code: 20500, Message: "invalid username or password"}
	ErrUserNotFound       = ErrorCode{Code: 20501, Message: "user not found"}
	ErrVerifyTooMany      = ErrorCode{Code: 20402, Message: "verify code req too frequent"}
)

// 题目相关错误
var (
	ErrProblemExists   = ErrorCode{Code: 30500, Message: "problem already exists"}
	ErrProblemNotFound = ErrorCode{Code: 30501, Message: "problem not found"}
	ErrTagExists       = ErrorCode{Code: 30502, Message: "tag already exists"}
	ErrNoTags          = ErrorCode{Code: 30503, Message: "no tag be found"}
)

// 文章相关错误
var (
	ErrAddDraft        = ErrorCode{Code: 40500, Message: "fail to create draft"}
	ErrUpdateDraft     = ErrorCode{Code: 40501, Message: "fail to update draft"}
	ErrSyncPublish     = ErrorCode{Code: 40502, Message: "fail to sync publish"}
	ErrWithdrawArt     = ErrorCode{Code: 40503, Message: "fail to withdraw article"}
	ErrArticleNotFound = ErrorCode{Code: 40504, Message: "article not found"}
)
