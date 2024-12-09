package domain

import "time"

type User struct {
	Id         uint64     `json:"id"`
	Name       string     `json:"name"`
	Password   string     `json:"password"`
	Email      string     `json:"email"`
	Phone      string     `json:"phone"`
	GithubID   string     `json:"github_id"`
	WeChatInfo WeChatInfo `json:"we_chat_info"`
	Birthday   time.Time  `json:"birthday"`
	Role       uint8      `json:"role"`
}
