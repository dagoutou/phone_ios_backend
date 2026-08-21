package model

// ProductPrice 商品价格表
type ProductPrice struct {
	ID     int     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name   string  `json:"name" gorm:"unique;not null;comment:商品名称"`
	Price  float64 `json:"price" gorm:"not null;comment:价格"`
}

func (p *ProductPrice) TableName() string {
	return "product_prices"
}
