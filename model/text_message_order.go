package model

import (
	"time"
)

// OrderType 订单类型
type OrderType string

const (
	OrderTypeTextMessage  OrderType = "text_message"
	OrderTypeManualMessage OrderType = "manual_message"
)

// MessageDetailType 消息详情类型
type MessageDetailType string

const (
	MessageDetailTypeTextMessage  MessageDetailType = "text_message"
	MessageDetailTypeManualMessage MessageDetailType = "manual_message"
)

// OrderStatus 订单状态
type OrderStatus string

const (
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusFailed    OrderStatus = "failed"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusSuccess   OrderStatus = "success"
	OrderStatusReviewFailed OrderStatus = "review_failed"
)

// TextMessageOrder 短信订单
type TextMessageOrder struct {
	ID        int        `json:"id"`
	OrderNo   string     `json:"order_no"`
	UserID    string     `json:"user_id"`
	OrderType OrderType  `json:"order_type"`
	Status    OrderStatus `json:"status"`
	PlanTime  *int64     `json:"plan_time"` // 定时发送时间（秒级时间戳）
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (t *TextMessageOrder) TableName() string {
	return "orders"
}

// TextMessageDetail 短信发送详情
type TextMessageDetail struct {
	ID              int       `json:"id"`
	APISendID       string    `json:"api_send_id"`
	UserID          string    `json:"user_id"`
	OrderNo         string    `json:"order_no"`
	Mobile          string    `json:"mobile"`
	Account         string    `json:"account"`
	Content         string    `json:"content"`
	Platform        string    `json:"platform"`
	PlatformAccount string    `json:"platform_account"`
	PlanTime        *time.Time `json:"plan_time"`
	CreateTime      *time.Time `json:"create_time"`
	Strip           int       `json:"strip"`
	Remarks         string    `json:"remarks"`
	Domain          string    `json:"domain"`
	Type            MessageDetailType `json:"type"`
	ImageURL        string    `json:"image_url"`
	RejectReason    string    `json:"reject_reason"`
	CreatedAt       time.Time `json:"created_at"`
}

func (t *TextMessageDetail) TableName() string {
	return "text_message_details"
}

// TextMessageReply 短信上行回复
type TextMessageReply struct {
	ID         int       `json:"id"`
	APISendID  string    `json:"api_send_id"`
	UserID     string    `json:"user_id"`
	OrderNo    string    `json:"order_no"`
	Domain     string    `json:"domain"`
	Content    string    `json:"content"`
	Mobile     string    `json:"mobile"`
	CreateTime *time.Time `json:"create_time"`
	Lang       string    `json:"lang"`
	ImageURL   string    `json:"image_url"` // 图片URL
	Type       MessageDetailType `json:"type"`
	CreatedAt  time.Time `json:"created_at"`
}

func (t *TextMessageReply) TableName() string {
	return "text_message_reply"
}
