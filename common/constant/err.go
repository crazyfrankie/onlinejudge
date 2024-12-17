package constant

type ErrorCode struct {
	Code    int
	Message string
}

// 通用错误
var (
	Success = ErrorCode{Code: 0, Message: "success"}

	ErrInternalServer = ErrorCode{Code: 10001, Message: "internal server errors"}
	ErrInvalidParams  = ErrorCode{Code: 10002, Message: "invalid parameters"}
	ErrUnauthorized   = ErrorCode{Code: 10003, Message: "unauthorized"}
	ErrForbidden      = ErrorCode{Code: 10004, Message: "forbidden"}
)

// 用户相关错误
var (
	ErrUserExists         = ErrorCode{Code: 20001, Message: "user already exists"}
	ErrUserNotFound       = ErrorCode{Code: 20002, Message: "user not found"}
	ErrInvalidCredentials = ErrorCode{Code: 20003, Message: "invalid username or password"}
	ErrInvalidToken       = ErrorCode{Code: 20004, Message: "invalid token"}
	ErrTokenExpired       = ErrorCode{Code: 20005, Message: "token expired"}
)

// 题目相关错误
var (
	ErrProblemExists   = ErrorCode{Code: 30001, Message: "problem already exists"}
	ErrProblemNotFound = ErrorCode{Code: 30002, Message: "problem not found"}
	ErrTagExists       = ErrorCode{Code: 30003, Message: "tag already exists"}
	ErrNoTags          = ErrorCode{Code: 30004, Message: "no tag be found"}
)

// 文章相关错误
var (
	ErrAddDraft        = ErrorCode{Code: 40001, Message: "fail to create draft"}
	ErrUpdateDraft     = ErrorCode{Code: 40002, Message: "fail to update draft"}
	ErrSyncPublish     = ErrorCode{Code: 40003, Message: "fail to sync publish"}
	ErrWithdrawArt     = ErrorCode{Code: 40004, Message: "fail to withdraw article"}
	ErrArticleNotFound = ErrorCode{Code: 40005, Message: "article not found"}
)
