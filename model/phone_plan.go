package model

type PhonePlan struct {
	PhoneCode string `json:"phoneCode"`
	PhoneName string `json:"phoneName"`
}

func (p *PhonePlan) TableName() string {
	return "phone_plans"
}
