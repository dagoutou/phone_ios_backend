package model

type BlackNumber struct {
	ID          int    `json:"id"`
	PhoneNumber string `json:"phone_number"`
}

func (p *BlackNumber) TableName() string {
	return "black_numbers"
}
