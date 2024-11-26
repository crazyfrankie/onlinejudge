package web

type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
	Err  string      `json:"err"`
}

func GetResponse(options ...func(*Response)) Response {
	resp := Response{
		Code: 200, // 默认状态
		Msg:  "",  // 默认消息
		Err:  "",  // 默认错误信息
	}
	for _, opt := range options {
		opt(&resp)
	}
	return resp
}

// 设置具体参数的函数

func WithStatus(code int) func(*Response) {
	return func(r *Response) {
		r.Code = code
	}
}

func WithData(data interface{}) func(*Response) {
	return func(r *Response) {
		r.Data = data
	}
}

func WithMsg(msg string) func(*Response) {
	return func(r *Response) {
		r.Msg = msg
	}
}
