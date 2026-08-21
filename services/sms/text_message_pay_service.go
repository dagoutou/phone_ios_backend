package sms

import (
	"phone_ios_backend/dao"
	"phone_ios_backend/logger"
	"phone_ios_backend/model"
	"phone_ios_backend/my_utils"
	"phone_ios_backend/services/mail"
	"encoding/json"
	"net/http"
	"time"
	
	"gorm.io/gorm"
)

// TextMessagePayParams 短信支付参数
type TextMessagePayParams struct {
	Content         string `json:"content" binding:"required"` // 短信内容
	Mobile          string `json:"mobile" binding:"required"`  // 收件人号码
	Account         string `json:"account"`                    // 发件人账号
	Houzhui         string `json:"houzhui"`                    // 短信后缀
	PlanTime        *int64 `json:"plan_time"`                  // 定时发送时间（秒级时间戳）
	MessageType     string `json:"message_type"`               // 消息类型：text_message 或 manual_message
	Platform        string `json:"platform"`                   // 手动消息时必填：传话平台
	PlatformAccount string `json:"platform_account"`           // 手动消息时必填：传话账号
}

// ProcessTextMessagePayment 处理短信支付
func ProcessTextMessagePayment(db *gorm.DB, userID, orderNo string, metadata string) error {
	// 解析元数据获取短信参数
	var params TextMessagePayParams
	if err := json.Unmarshal([]byte(metadata), &params); err != nil {
		return logger.NewAppError(http.StatusBadRequest, "Invalid metadata format: "+err.Error())
	}
	
	// 创建短信订单
	smsOrder := &model.TextMessageOrder{
		OrderNo:   orderNo,
		UserID:    userID,
		OrderType: model.OrderTypeTextMessage,
		Status:    model.OrderStatusPaid,
		PlanTime:  params.PlanTime,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// 开启事务
	return db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建短信订单
		if err := tx.Create(smsOrder).Error; err != nil {
			return err
		}
		
		// 2. 调用短信发送服务
		smsService := NewSMSService()
		smsResp, err := smsService.SendTextMessage(SendSMSParams{
			Content:      params.Content,
			Mobile:       params.Mobile,
			Account:      params.Account,
			PlanTime:     params.PlanTime,
			SMSAutograph: "102846",
			Houzhui:      params.Houzhui,
		})
		if err != nil {
			logger.Logger.Errorf("Failed to send SMS: %v", err)
			// 更新订单状态为失败
			tx.Model(&model.TextMessageOrder{}).
				Where("order_no = ?", orderNo).
				Update("status", model.OrderStatusFailed)
			return err
		}
		
		// 3. 检查短信发送结果
		if !smsResp.Bool {
			logger.Logger.Errorf("SMS send failed: %s", smsResp.Msg)
			// 更新订单状态为失败
			tx.Model(&model.TextMessageOrder{}).
				Where("order_no = ?", orderNo).
				Update("status", model.OrderStatusFailed)
			return logger.NewAppError(http.StatusInternalServerError, "SMS send failed: "+smsResp.Msg)
		}
		
		// 4. 创建短信详情记录（类型为 text_message）
		detail := ResponseToDetail(smsResp, userID, orderNo, params.Content, model.MessageDetailTypeTextMessage)
		if detail != nil {
			if err := tx.Create(detail).Error; err != nil {
				logger.Logger.Errorf("Failed to create SMS detail: %v", err)
				// 这里不需要回滚，因为短信已经发送成功
			}
		}
		
		logger.Logger.Infof("SMS payment processed successfully for order: %s", orderNo)
		return nil
	})
}

// ProcessManualMessagePayment 处理手动短信订单（状态为 processing）
func ProcessManualMessagePayment(db *gorm.DB, userID, orderNo string, metadata string) error {
	// 解析元数据获取短信参数
	var params TextMessagePayParams
	if err := json.Unmarshal([]byte(metadata), &params); err != nil {
		return logger.NewAppError(http.StatusBadRequest, "Invalid metadata format: "+err.Error())
	}
	
	// 创建手动短信订单（状态为 processing）
	smsOrder := &model.TextMessageOrder{
		OrderNo:   orderNo,
		UserID:    userID,
		OrderType: model.OrderTypeManualMessage,
		Status:    model.OrderStatusPaid,
		PlanTime:  params.PlanTime,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// 开启事务
	return db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建短信订单
		if err := tx.Create(smsOrder).Error; err != nil {
			return err
		}
		
		// 2. 创建短信详情记录（类型为 manual_message，备注为 processing）
		detail := &model.TextMessageDetail{
			UserID:          userID,
			OrderNo:         orderNo,
			Mobile:          params.Mobile,
			Account:         params.Account,
			Content:         params.Content,
			Platform:        params.Platform,
			PlatformAccount: params.PlatformAccount,
			Remarks:         "reviewing",
			Type:            model.MessageDetailTypeManualMessage,
			CreatedAt:       time.Now(),
		}
		if params.PlanTime != nil {
			t := time.Unix(*params.PlanTime, 0)
			detail.PlanTime = &t
		}
		if err := tx.Create(detail).Error; err != nil {
			logger.Logger.Errorf("Failed to create manual message detail: %v", err)
			return err
		}
		
		// 3. 发送邮件通知（异步，不阻塞主流程）
		go func() {
			mailService := mail.NewMailService()
			if mailErr := mailService.SendManualMessageOrderNotification(orderNo, userID, params.Platform, params.PlatformAccount, params.Content); mailErr != nil {
				logger.Logger.Errorf("Failed to send manual message order notification email: %v", mailErr)
			}
		}()
		
		logger.Logger.Infof("Manual SMS payment processed successfully for order: %s", orderNo)
		return nil
	})
}

// GetTextMessageOrders 获取用户短信订单列表
func GetTextMessageOrders(userID string) ([]model.TextMessageOrder, error) {
	db := dao.NewDB()
	return db.GetTextMessageOrdersByUserID(userID)
}

// GetTextMessageDetails 获取用户短信详情列表，detailType 为空时不过滤类型
func GetTextMessageDetails(userID string, detailType model.MessageDetailType) ([]model.TextMessageDetail, error) {
	db := dao.NewDB()
	return db.GetTextMessageDetailsByUserID(userID, detailType)
}

// GetTextMessageOrderDetail 获取短信订单详情
func GetTextMessageOrderDetail(orderNo string) (*model.TextMessageDetail, error) {
	db := dao.NewDB()
	return db.GetTextMessageDetailByOrderNo(orderNo)
}

// GenerateTextMessageOrderNo 生成短信订单号
func GenerateTextMessageOrderNo() string {
	return my_utils.GenerateOrderID("SMS")
}
