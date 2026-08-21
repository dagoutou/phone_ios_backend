package model

import "time"

type User struct {
	ID         int       `json:"id"`
	UserID     string    `json:"user_id"`
	CreateTime time.Time `json:"create_time"`
}

func (u *User) TableName() string {
	return "users"
}
