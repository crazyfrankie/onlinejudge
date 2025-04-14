package web

type SendCodeReq struct {
	Phone string `json:"phone"`
	Biz   string `json:"biz"`
}

type VerifyCodeReq struct {
	Phone string `json:"phone" validate:"required,len=11"`
	Code  string `json:"code"`
	Role  string `json:"role"`
	Biz   string `json:"biz"`
}

type LoginReq struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type UpdatePwdReq struct {
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirmPassword" validate:"eqfield=Password"`
}

type UpdateInfoReq struct {
	Email    string `json:"email" validate:"required,email"`
	Birthday string `json:"birthday"`
	Name     string `json:"name" validate:"required,min=3,max=20"`
}

type UpdateRoleReq struct {
	Role uint8 `json:"role"`
}
