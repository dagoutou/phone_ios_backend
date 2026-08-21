package model

type PriceMenu struct {
	ID       int     `json:"id"`
	Menu     string  `json:"menu"`
	Price    float64 `json:"price"`       // 价格
	Amount   float64 `json:"real_amount"` // 实际到账金额
	Describe string  `json:"describe"`
}

func (u *PriceMenu) TableName() string {
	return "price_menus"
}
