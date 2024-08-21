package domain

type User struct {
	Id       uint64
	Name     string
	Password string
	Email    string
	Phone    string
	Role     uint8
}
