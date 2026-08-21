package api

import (
	"net/http"
	"phone_ios_backend/common"
	"phone_ios_backend/connection"
	"phone_ios_backend/dao"
	"phone_ios_backend/logger"
	"phone_ios_backend/model"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SMSUpstreamRequest 短信上行回复请求
type SMSUpstreamRequest struct {
	Domain     string `json:"domain"`
	Content    string `json:"content"`
	Mobile     string `json:"mobile"`
	APISendID  string `json:"api_send_id" binding:"required"`
	CreateTime string `json:"create_time"` // 格式: "2022-08-04 17:34:51"
	Lang       string `json:"lang"`
}

// PullSMSUpstream 接收短信上行回复
func PullSMSUpstream(ctx *gin.Context) {
	var req SMSUpstreamRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Logger.Errorf("Failed to bind SMS upstream request: %v", err)
		ctx.JSON(http.StatusOK, gin.H{
			"code":    http.StatusBadRequest,
			"message": "Invalid request parameters",
		})
		return
	}

	logger.Logger.Infof("Received SMS upstream: api_send_id=%s, mobile=%s, content=%s",
		req.APISendID, req.Mobile, req.Content)

	db := dao.NewDB()

	// 根据api_send_id查询原短信详情
	detail, err := db.GetTextMessageDetailByAPISendID(req.APISendID)
	if err != nil {
		logger.Logger.Errorf("Failed to get SMS detail by api_send_id %s: %v", req.APISendID, err)
		ctx.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": "success",
		})
		return
	}

	// 解析创建时间
	var createTime *time.Time
	if req.CreateTime != "" {
		t, err := time.Parse("2006-01-02 15:04:05", req.CreateTime)
		if err == nil {
			createTime = &t
		}
	}

	// 创建短信回复记录
	reply := &model.TextMessageReply{
		APISendID:  req.APISendID,
		UserID:     detail.UserID,
		OrderNo:    detail.OrderNo,
		Domain:     req.Domain,
		Content:    req.Content,
		Mobile:     req.Mobile,
		CreateTime: createTime,
		Lang:       req.Lang,
		Type:       model.MessageDetailTypeTextMessage,
		CreatedAt:  time.Now(),
	}

	// 使用事务处理
	if err := connection.DbConnection.Transaction(func(tx *gorm.DB) error {
		// 1. 创建短信回复记录
		if err := tx.Create(reply).Error; err != nil {
			logger.Logger.Errorf("Failed to create SMS reply: %v", err)
			return err
		}

		logger.Logger.Infof("SMS upstream processed successfully: api_send_id=%s", req.APISendID)
		return nil
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to process SMS upstream",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "success",
	})
}

// SMSStatusRequest 短信状态推送请求
type SMSStatusRequest struct {
	Domain    string `json:"domain"`
	SendID    string `json:"send_id"`     // 发送编号，暂时没用可以忽略
	Remarks   string `json:"remarks"`     // 发送失败的提示
	Account   string `json:"account"`     // 发送人账号
	APISendID string `json:"api_send_id"` // 发送的唯一id，和发短信时返回的对应
	Lang      string `json:"lang"`
}

// PullSMSStatus 接收短信状态推送（失败通知）
func PullSMSStatus(ctx *gin.Context) {
	var req SMSStatusRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Logger.Errorf("Failed to bind SMS status request: %v", err)
		ctx.JSON(http.StatusOK, gin.H{
			"code":    http.StatusBadRequest,
			"message": "Invalid request parameters",
		})
		return
	}

	logger.Logger.Infof("Received SMS status notification: api_send_id=%s, remarks=%s, account=%s",
		req.APISendID, req.Remarks, req.Account)

	db := dao.NewDB()

	// 根据 api_send_id 查询短信详情
	detail, err := db.GetTextMessageDetailByAPISendID(req.APISendID)
	if err != nil {
		logger.Logger.Errorf("Failed to get SMS detail by api_send_id %s: %v", req.APISendID, err)
		ctx.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": "success",
		})
		return
	}

	// 查询订单，若已 failed 则不再退还（幂等）
	order, err := db.GetTextMessageOrderByOrderNo(detail.OrderNo)
	if err != nil {
		logger.Logger.Errorf("Failed to get SMS order by order_no %s: %v", detail.OrderNo, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to get order",
		})
		return
	}
	if order.Status == model.OrderStatusFailed {
		logger.Logger.Infof("Order %s already failed, skip refund", detail.OrderNo)
		ctx.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": "success",
		})
		return
	}

	// 查询 text_message 价格
	// productPriceDAO := dao.NewProductPriceDAO()
	// productPrice, err := productPriceDAO.GetProductPriceByName(string(model.ProductTypeTextMessage))
	// if err != nil {
	// 	logger.Logger.Errorf("Failed to get text_message price: %v", err)
	// 	ctx.JSON(http.StatusInternalServerError, gin.H{
	// 		"code":    http.StatusInternalServerError,
	// 		"message": "Failed to get product price",
	// 	})
	// 	return
	// }

	// 事务：退还余额 + 订单状态置为 failed + 详情备注置为 failed
	if err := connection.DbConnection.Transaction(func(tx *gorm.DB) error {
		// if err := tx.Model(&model.UserBalance{}).
		// 	Where("user_id = ?", order.UserID).
		// 	Update("balance", gorm.Expr("balance + ?", productPrice.Price)).Error; err != nil {
		// 	return err
		// }
		if err := tx.Model(&model.TextMessageOrder{}).
			Where("order_no = ?", detail.OrderNo).
			Update("status", model.OrderStatusFailed).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.TextMessageDetail{}).
			Where("api_send_id = ?", req.APISendID).
			Update("remarks", "failed").Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		logger.Logger.Errorf("Failed to process SMS refund transaction: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to refund",
		})
		return
	}

	// logger.Logger.Infof("SMS refund processed: order_no=%s, amount=%.2f", detail.OrderNo, productPrice.Price)

	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "success",
	})
}

// GetSMSReplies 获取用户的短信回复列表
func GetSMSReplies(ctx *gin.Context) {
	userID := ctx.Query("user_id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, common.RequestResp{
			Code:    http.StatusBadRequest,
			Message: "user_id is required",
		})
		return
	}

	messageType := model.MessageDetailType(ctx.Query("message_type"))
	if messageType == "" {
		ctx.JSON(http.StatusBadRequest, common.RequestResp{
			Code:    http.StatusBadRequest,
			Message: "message_type is required",
		})
		return
	}

	db := dao.NewDB()
	replies, err := db.GetTextMessageRepliesByUserID(userID, messageType)
	if err != nil {
		logger.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.RequestResp{
		Code:    http.StatusOK,
		Data:    replies,
		Message: "success",
	})
}

// GetSMSRepliesByOrderNo 获取订单的短信回复列表
func GetSMSRepliesByOrderNo(ctx *gin.Context) {
	orderNo := ctx.Query("order_no")
	if orderNo == "" {
		ctx.JSON(http.StatusBadRequest, common.RequestResp{
			Code:    http.StatusBadRequest,
			Message: "order_no is required",
		})
		return
	}

	var replies []model.TextMessageReply
	if err := connection.DbConnection.Where("order_no = ?", orderNo).Order("created_at desc").Find(&replies).Error; err != nil {
		logger.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.RequestResp{
		Code:    http.StatusOK,
		Data:    replies,
		Message: "success",
	})
}
