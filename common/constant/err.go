package constant

type ErrorCode struct {
	Code    int32
	Message string
}

// 错误码采用五位设计
// 第一位代表错误类型 4 代表客户端错误、 5 代表服务端错误
// 后两位代表模块
// 最后两位代表具体错误，从 00 开始递增

// 通用错误
var (
	Success = ErrorCode{Code: 00000, Message: "success"}

	ErrTooManyRequests = ErrorCode{Code: 40000, Message: "too many requests"}
	ErrInternalServer  = ErrorCode{Code: 50001, Message: "internal server errors"}
)

// 用户相关错误
var (
	ErrUserInvalidParams  = ErrorCode{Code: 40100, Message: "invalid parameters"}
	ErrUserInvalidToken   = ErrorCode{Code: 40101, Message: "invalid token"}
	ErrUserTokenExpired   = ErrorCode{Code: 40102, Message: "token expired"}
	ErrUserLoginYet       = ErrorCode{Code: 40103, Message: "have not logged in yet"}
	ErrUserSessExpired    = ErrorCode{Code: 40104, Message: "session expired"}
	ErrUserUnauthorized   = ErrorCode{Code: 40105, Message: "unauthorized"}
	ErrUserForbidden      = ErrorCode{Code: 40106, Message: "forbidden"}
	ErrUserNotFound       = ErrorCode{Code: 40107, Message: "user not found"}
	ErrVerifyTooMany      = ErrorCode{Code: 40108, Message: "verify code req too frequent"}
	ErrInvalidCredentials = ErrorCode{Code: 40109, Message: "invalid username or password"}
	ErrUserInternalServer = ErrorCode{Code: 50103, Message: "internal server errors"}
)

// 题目相关错误
var (
	ErrProblemNotFound       = ErrorCode{Code: 40200, Message: "problem not found"}
	ErrProblemExists         = ErrorCode{Code: 40201, Message: "problem already exists"}
	ErrProblemTagExists      = ErrorCode{Code: 40202, Message: "tag already exists"}
	ErrProblemNoTags         = ErrorCode{Code: 40203, Message: "no tag be found"}
	ErrProblemInternalServer = ErrorCode{Code: 50204, Message: "internal server error"}
)

// 文章相关错误
var (
	ErrArticleInvalidParams  = ErrorCode{Code: 40300, Message: "invalid parameters"}
	ErrArticleNotFound       = ErrorCode{Code: 40301, Message: "article not found"}
	ErrArtilceForbidden      = ErrorCode{Code: 40302, Message: "forbidden"}
	ErrArticleInternalServer = ErrorCode{Code: 50303, Message: "internal server error"}
)

// 交互系统相关错误
var (
	ErrInteractiveInternalServer = ErrorCode{Code: 50400, Message: "internal server error"}
)

// 验证码系统相关错误
var (
	ErrCodeInternalServer = ErrorCode{Code: 50500, Message: "internal server error"}
)
