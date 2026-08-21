package model

import "time"

type UserRealNameAuthentication struct {
	ID         int       `json:"id"`
	UserID     string    `json:"user_id"`
	CreateTime time.Time `json:"create_time"`
}

func (p *UserRealNameAuthentication) TableName() string {
	return "user_real_name_authentication"
}
