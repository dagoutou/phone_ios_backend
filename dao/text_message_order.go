package dao

import (
	"phone_ios_backend/model"

	"gorm.io/gorm"
)

// CreateTextMessageOrder 创建短信订单
func (d *DBConnection) CreateTextMessageOrder(order *model.TextMessageOrder) error {
	return d.db.Create(order).Error
}

// GetTextMessageOrderByID 根据ID获取短信订单
func (d *DBConnection) GetTextMessageOrderByID(id int) (*model.TextMessageOrder, error) {
	var order model.TextMessageOrder
	err := d.db.First(&order, id).Error
	return &order, err
}

// GetTextMessageOrderByOrderNo 根据订单号获取短信订单
func (d *DBConnection) GetTextMessageOrderByOrderNo(orderNo string) (*model.TextMessageOrder, error) {
	var order model.TextMessageOrder
	err := d.db.Where("order_no = ?", orderNo).First(&order).Error
	return &order, err
}

// UpdateTextMessageOrder 更新短信订单
func (d *DBConnection) UpdateTextMessageOrder(order *model.TextMessageOrder) error {
	return d.db.Save(order).Error
}

// GetTextMessageOrdersByUserID 获取用户的所有短信订单
func (d *DBConnection) GetTextMessageOrdersByUserID(userID string) ([]model.TextMessageOrder, error) {
	var orders []model.TextMessageOrder
	err := d.db.Where("user_id = ?", userID).Order("created_at desc").Find(&orders).Error
	return orders, err
}

// CreateTextMessageDetail 创建短信详情
func (d *DBConnection) CreateTextMessageDetail(detail *model.TextMessageDetail) error {
	return d.db.Create(detail).Error
}

// GetTextMessageDetailByOrderNo 根据订单号获取短信详情
func (d *DBConnection) GetTextMessageDetailByOrderNo(orderNo string) (*model.TextMessageDetail, error) {
	var detail model.TextMessageDetail
	err := d.db.Where("order_no = ?", orderNo).First(&detail).Error
	return &detail, err
}

// GetTextMessageDetailsByUserID 获取用户的所有短信详情，detailType 为空时不过滤类型
func (d *DBConnection) GetTextMessageDetailsByUserID(userID string, detailType model.MessageDetailType) ([]model.TextMessageDetail, error) {
	var details []model.TextMessageDetail
	query := d.db.Where("user_id = ?", userID)
	if detailType != "" {
		query = query.Where("type = ?", detailType)
	}
	err := query.Order("created_at desc").Find(&details).Error
	return details, err
}

// CreateTextMessageReply 创建短信上行回复
func (d *DBConnection) CreateTextMessageReply(reply *model.TextMessageReply) error {
	return d.db.Create(reply).Error
}

// GetTextMessageReplyByAPISendID 根据API发送ID获取回复
func (d *DBConnection) GetTextMessageReplyByAPISendID(apiSendID string) (*model.TextMessageReply, error) {
	var reply model.TextMessageReply
	err := d.db.Where("api_send_id = ?", apiSendID).First(&reply).Error
	return &reply, err
}

// GetTextMessageRepliesByUserID 获取用户的所有短信回复
func (d *DBConnection) GetTextMessageRepliesByUserID(userID string, messageType model.MessageDetailType) ([]model.TextMessageReply, error) {
	var replies []model.TextMessageReply
	err := d.db.Where("user_id = ? AND type = ?", userID, messageType).Order("created_at desc").Find(&replies).Error
	return replies, err
}

// UpdateTextMessageDetailRemarks 更新短信详情备注
func (d *DBConnection) UpdateTextMessageDetailRemarks(orderNo string, remarks string) error {
	return d.db.Model(&model.TextMessageDetail{}).
		Where("order_no = ?", orderNo).
		Update("remarks", remarks).Error
}

// RejectTextMessageOrder 驳回订单:订单状态置为 review_failed,
// TextMessageDetail 的 remarks 置为 review_failed,reject_reason 写入驳回原因,
// 同时退还用户余额(refundAmount > 0 时)
func (d *DBConnection) RejectTextMessageOrder(orderNo, userID, rejectReason string, refundAmount float64) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		// if refundAmount > 0 {
		// 	if err := tx.Model(&model.UserBalance{}).
		// 		Where("user_id = ?", userID).
		// 		Update("balance", gorm.Expr("balance + ?", refundAmount)).Error; err != nil {
		// 		return err
		// 	}
		// }
		if err := tx.Model(&model.TextMessageOrder{}).
			Where("order_no = ? AND user_id = ?", orderNo, userID).
			Update("status", model.OrderStatusFailed).Error; err != nil {
			return err
		}
		updates := map[string]interface{}{
			"remarks":       string(model.OrderStatusFailed),
			"reject_reason": rejectReason,
		}
		if err := tx.Model(&model.TextMessageDetail{}).
			Where("order_no = ? AND user_id = ?", orderNo, userID).
			Updates(updates).Error; err != nil {
			return err
		}
		return nil
	})
}

// GetTextMessageDetailByAPISendID 根据API发送ID获取短信详情
func (d *DBConnection) GetTextMessageDetailByAPISendID(apiSendID string) (*model.TextMessageDetail, error) {
	var detail model.TextMessageDetail
	err := d.db.Where("api_send_id = ?", apiSendID).First(&detail).Error
	return &detail, err
}
