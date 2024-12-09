package constant

type ErrorCode struct {
	Code    int
	Message string
}

var (
	Success = ErrorCode{Code: 0, Message: "success"}

	// 通用错误
	ErrInternalServer = ErrorCode{Code: 10001, Message: "internal server error"}
	ErrInvalidParams  = ErrorCode{Code: 10002, Message: "invalid parameters"}
	ErrUnauthorized   = ErrorCode{Code: 10003, Message: "unauthorized"}
	ErrForbidden      = ErrorCode{Code: 10004, Message: "forbidden"}

	// 用户相关错误
	ErrUserExists         = ErrorCode{Code: 20001, Message: "user already exists"}
	ErrUserNotFound       = ErrorCode{Code: 20002, Message: "user not found"}
	ErrInvalidCredentials = ErrorCode{Code: 20003, Message: "invalid username or password"}
	ErrInvalidToken       = ErrorCode{Code: 20004, Message: "invalid token"}
	ErrTokenExpired       = ErrorCode{Code: 20005, Message: "token expired"}

	// 题目相关错误
	ErrProblemExists   = ErrorCode{Code: 30001, Message: "problem already exists"}
	ErrProblemNotFound = ErrorCode{Code: 30002, Message: "problem not found"}
	ErrTagExists       = ErrorCode{Code: 30003, Message: "tag already exists"}
	ErrNoTags          = ErrorCode{Code: 30004, Message: "no tag be found"}
)
