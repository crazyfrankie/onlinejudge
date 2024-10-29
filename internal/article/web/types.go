package web

type Response struct {
	Status int    `json:"status"`
	Data   uint64 `json:"data"`
	Msg    string `json:"msg"`
}

type Result[T any] struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
	Data T      `json:"data"`
}

func GetResponse(options ...func(*Response)) Response {
	resp := Response{
		Status: 200, // 默认状态
		Msg:    "",  // 默认消息
	}
	for _, opt := range options {
		opt(&resp)
	}
	return resp
}

// 设置具体参数的函数

func WithStatus(status int) func(*Response) {
	return func(r *Response) {
		r.Status = status
	}
}

func WithData(data uint64) func(*Response) {
	return func(r *Response) {
		r.Data = data
	}
}

func WithMsg(msg string) func(*Response) {
	return func(r *Response) {
		r.Msg = msg
	}
}
