package constant

import "net/http"

type ErrorCode struct {
	Code    int32
	Message string
	Status  int
}

// 通用错误
var (
	Success = ErrorCode{Code: 00000, Message: "success"}

	ErrInvalidParams   = ErrorCode{Code: 00400, Message: "invalid parameters", Status: http.StatusBadRequest}
	ErrTooManyRequests = ErrorCode{Code: 00401, Message: "too many requests", Status: http.StatusTooManyRequests}
	ErrInternalServer  = ErrorCode{Code: 00500, Message: "internal server errors", Status: http.StatusInternalServerError}
)

// 身份校验错误
var (
	ErrInvalidToken = ErrorCode{Code: 10400, Message: "invalid token", Status: http.StatusUnauthorized}
	ErrTokenExpired = ErrorCode{Code: 10401, Message: "token expired", Status: http.StatusUnauthorized}
	ErrLoginYet     = ErrorCode{Code: 10402, Message: "have not logged in yet", Status: http.StatusUnauthorized}
	ErrSessExpired  = ErrorCode{Code: 10403, Message: "session expired", Status: http.StatusUnauthorized}
	ErrUnauthorized = ErrorCode{Code: 10404, Message: "unauthorized", Status: http.StatusUnauthorized}
	ErrForbidden    = ErrorCode{Code: 10405, Message: "forbidden", Status: http.StatusForbidden}
)

// 用户相关错误
var (
	ErrUserNotFound       = ErrorCode{Code: 20400, Message: "user not found", Status: http.StatusNotFound}
	ErrVerifyTooMany      = ErrorCode{Code: 20401, Message: "verify code req too frequent", Status: http.StatusTooManyRequests}
	ErrInvalidCredentials = ErrorCode{Code: 20500, Message: "invalid username or password", Status: http.StatusUnauthorized}
)

// 题目相关错误
var (
	ErrProblemNotFound = ErrorCode{Code: 30400, Message: "problem not found", Status: http.StatusNotFound}
	ErrProblemExists   = ErrorCode{Code: 30501, Message: "problem already exists", Status: http.StatusConflict}
	ErrTagExists       = ErrorCode{Code: 30502, Message: "tag already exists", Status: http.StatusConflict}
	ErrNoTags          = ErrorCode{Code: 30503, Message: "no tag be found", Status: http.StatusNotFound}
)

// 文章相关错误
var (
	ErrArticleNotFound = ErrorCode{Code: 40504, Message: "article not found", Status: http.StatusNotFound}
)
