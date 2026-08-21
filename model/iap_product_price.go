package model

import "time"

// IAPProductPrice Apple IAP 内购商品配置
type IAPProductPrice struct {
	ID         int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name       string    `json:"name" gorm:"size:100;unique;not null;comment:内部名称(如credits6)"`
	ProductID  string    `json:"product_id" gorm:"size:255;unique;not null;comment:App Store product_id"`
	Price      float64   `json:"price" gorm:"not null;comment:人民币价格"`
	Credits    int       `json:"credits" gorm:"not null;comment:赠送积分"`
	CreateTime time.Time `json:"create_time" gorm:"autoCreateTime"`
	UpdateTime time.Time `json:"update_time" gorm:"autoUpdateTime"`
}

func (p *IAPProductPrice) TableName() string {
	return "iap_product_prices"
}
