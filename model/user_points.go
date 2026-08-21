package model

import "time"

// UserPoints 用户积分余额
type UserPoints struct {
	ID         int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID     string    `json:"user_id" gorm:"size:255;uniqueIndex:idx_user_app,priority:1"`
	AppName    string    `json:"app_name" gorm:"size:255;uniqueIndex:idx_user_app,priority:2"`
	Balance    int       `json:"balance"`
	CreateTime time.Time `json:"create_time" gorm:"autoCreateTime"`
	UpdateTime time.Time `json:"update_time" gorm:"autoUpdateTime"`
}

func (p *UserPoints) TableName() string {
	return "user_points"
}

const (
	PointTypeRecharge = "recharge"
	PointTypeConsume  = "consume"
	PointTypeRefund   = "refund"

	PointSourceIAP    = "iap"
	PointSourceAlipay = "alipay"
	PointSourceWechat = "wechat"
	PointSourceManual = "manual"
	PointSourceAdmin  = "admin"
)

// PointTransaction 积分变动明细
type PointTransaction struct {
	ID           int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID       string    `json:"user_id" gorm:"size:255;index:idx_user_id"`
	AppName      string    `json:"app_name" gorm:"size:255;index:idx_app_name"`
	Type         string    `json:"type" gorm:"size:50"`
	Amount       int       `json:"amount"`
	BalanceAfter int       `json:"balance_after"`
	Source       string    `json:"source" gorm:"size:50"`
	RefOrderNo   string    `json:"ref_order_no" gorm:"size:255"`
	ProductID    string    `json:"product_id" gorm:"size:255"`
	Remark       string    `json:"remark" gorm:"size:500"`
	CreateTime   time.Time `json:"create_time" gorm:"autoCreateTime"`
}

func (p *PointTransaction) TableName() string {
	return "point_transactions"
}
