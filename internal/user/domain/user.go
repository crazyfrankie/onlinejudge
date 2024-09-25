package domain

import "time"

type User struct {
	Id         uint64
	Name       string
	Password   string
	Email      string
	Phone      string
	WeChatInfo WeChatInfo
	AboutMe    string
	Birthday   time.Time
	Role       uint8
}
