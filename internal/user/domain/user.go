package domain

import "time"

type User struct {
	Id         uint64
	Name       string
	Password   string
	Email      string
	Phone      string
	GithubID   string
	WeChatInfo WeChatInfo
	Birthday   time.Time
	Role       uint8
}
