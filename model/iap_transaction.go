package model

import "time"

// IAPTransaction Apple IAP 交易记录
type IAPTransaction struct {
	ID                    int       `json:"id" gorm:"primaryKey;autoIncrement"`
	TransactionID         string    `json:"transaction_id" gorm:"size:255;uniqueIndex:idx_transaction_id"`
	OriginalTransactionID string    `json:"original_transaction_id" gorm:"size:255"`
	UserID                string    `json:"user_id" gorm:"size:255;index:idx_user_id"`
	AppName               string    `json:"app_name" gorm:"size:255;index:idx_app_name"`
	ProductID             string    `json:"product_id" gorm:"size:255"`
	Amount                float64   `json:"amount"`
	Credits               int       `json:"credits"`
	Environment           string    `json:"environment" gorm:"size:50"`
	ReceiptHash           string    `json:"receipt_hash" gorm:"size:255"`
	PayTime               time.Time `json:"pay_time"`
	CreateTime            time.Time `json:"create_time" gorm:"autoCreateTime"`
}

func (p *IAPTransaction) TableName() string {
	return "iap_transactions"
}
