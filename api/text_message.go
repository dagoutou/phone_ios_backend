package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"phone_ios_backend/common"
	"phone_ios_backend/connection"
	"phone_ios_backend/dao"
	"phone_ios_backend/logger"
	"phone_ios_backend/model"
	"phone_ios_backend/my_utils"
	"phone_ios_backend/services/sms"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GetTextMessageOrders 获取用户的短信订单列表
func GetTextMessageOrders(ctx *gin.Context) {
	userID := ctx.Query("user_id")
	if userID == "" {
		err := logger.NewAppError(http.StatusBadRequest, "user_id is required")
		logger.HandleError(ctx, err)
		return
	}

	orders, err := sms.GetTextMessageOrders(userID)
	if err != nil {
		logger.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.RequestResp{
		Code:    http.StatusOK,
		Data:    orders,
		Message: "success",
	})
}

// GetTextMessageDetails 获取用户的短信详情列表
func GetTextMessageDetails(ctx *gin.Context) {
	userID := ctx.Query("user_id")
	if userID == "" {
		err := logger.NewAppError(http.StatusBadRequest, "user_id is required")
		logger.HandleError(ctx, err)
		return
	}

	detailType := model.MessageDetailType(ctx.Query("message_type"))
	if detailType == "" {
		err := logger.NewAppError(http.StatusBadRequest, "message_type is required")
		logger.HandleError(ctx, err)
		return
	}

	details, err := sms.GetTextMessageDetails(userID, detailType)
	if err != nil {
		logger.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.RequestResp{
		Code:    http.StatusOK,
		Data:    details,
		Message: "success",
	})
}

// GetTextMessageOrderDetail 获取短信订单详情
func GetTextMessageOrderDetail(ctx *gin.Context) {
	orderNo := ctx.Query("order_no")
	if orderNo == "" {
		err := logger.NewAppError(http.StatusBadRequest, "order_no is required")
		logger.HandleError(ctx, err)
		return
	}

	detail, err := sms.GetTextMessageOrderDetail(orderNo)
	if err != nil {
		logger.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.RequestResp{
		Code:    http.StatusOK,
		Data:    detail,
		Message: "success",
	})
}

const (
	replyImageSaveDir   = "/www/server/nginx/conf/images/reply_images"
	replyImageURLPrefix = "https://phonewechat.addomi.cn/customer-images/images/reply_images"
	replyImageMaxSize   = 10 << 20 // 10MB
)

// UploadTextMessageReplyImage 上传短信回复图片并创建回复记录
func UploadTextMessageReplyImage(ctx *gin.Context) {
	// 解析表单数据
	userID := ctx.PostForm("user_id")
	orderNo := ctx.PostForm("order_no")

	if userID == "" || orderNo == "" {
		err := logger.NewAppError(http.StatusBadRequest, "user_id and order_no are required")
		logger.HandleError(ctx, err)
		return
	}

	// 获取上传文件
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		err := logger.NewAppError(http.StatusBadRequest, "file is required")
		logger.HandleError(ctx, err)
		return
	}

	// 校验文件大小
	if fileHeader.Size > replyImageMaxSize {
		err := logger.NewAppError(http.StatusBadRequest, "file size exceeds 10MB limit")
		logger.HandleError(ctx, err)
		return
	}

	// 校验扩展名
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		err := logger.NewAppError(http.StatusBadRequest, "only png/jpg/jpeg images are allowed")
		logger.HandleError(ctx, err)
		return
	}

	// 确保目录存在
	if err := os.MkdirAll(replyImageSaveDir, 0o755); err != nil {
		logger.Logger.Errorf("Failed to create reply image dir: %v", err)
		err := logger.NewAppError(http.StatusInternalServerError, "failed to prepare upload directory")
		logger.HandleError(ctx, err)
		return
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), my_utils.GenerateRandomString(8), ext)

	// 保存文件到服务器
	savePath := filepath.Join(replyImageSaveDir, filename)
	if err := ctx.SaveUploadedFile(fileHeader, savePath); err != nil {
		logger.Logger.Errorf("Failed to save reply image: %v", err)
		err := logger.NewAppError(http.StatusInternalServerError, "failed to save image")
		logger.HandleError(ctx, err)
		return
	}

	// 构建图片URL
	imageURL := fmt.Sprintf("%s/%s", replyImageURLPrefix, filename)

	// 创建短信回复记录
	db := dao.NewDB()
	n := time.Now()
	reply := &model.TextMessageReply{
		UserID:     userID,
		OrderNo:    orderNo,
		ImageURL:   imageURL,
		Type:       model.MessageDetailTypeManualMessage,
		CreateTime: &n, // 使用当前时间
		CreatedAt:  n,
	}

	if err := db.CreateTextMessageReply(reply); err != nil {
		logger.Logger.Errorf("Failed to create text message reply: %v", err)
		// 即使数据库创建失败，图片已经上传成功，不需要删除图片
		err = logger.NewAppError(http.StatusInternalServerError, "failed to create reply record: "+err.Error())
		logger.HandleError(ctx, err)
		return
	}

	logger.Logger.Infof("Text message reply created successfully: order_no=%s, image_url=%s", orderNo, imageURL)

	ctx.JSON(http.StatusOK, common.RequestResp{
		Code: http.StatusOK,
		Data: gin.H{
			"reply_id":   reply.ID,
			"image_url":  imageURL,
			"created_at": reply.CreatedAt,
		},
		Message: "success",
	})
}

// CompleteTextMessageOrder 完成短信订单：上传图片并将订单状态置为 success，
// 同时把图片 URL 写入对应 TextMessageDetail.image_url
func CompleteTextMessageOrder(ctx *gin.Context) {
	userID := ctx.PostForm("user_id")
	orderNo := ctx.PostForm("order_no")
	if userID == "" || orderNo == "" {
		err := logger.NewAppError(http.StatusBadRequest, "user_id and order_no are required")
		logger.HandleError(ctx, err)
		return
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		err := logger.NewAppError(http.StatusBadRequest, "file is required")
		logger.HandleError(ctx, err)
		return
	}
	if fileHeader.Size > replyImageMaxSize {
		err := logger.NewAppError(http.StatusBadRequest, "file size exceeds 10MB limit")
		logger.HandleError(ctx, err)
		return
	}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		err := logger.NewAppError(http.StatusBadRequest, "only png/jpg/jpeg images are allowed")
		logger.HandleError(ctx, err)
		return
	}

	// 校验订单存在且属于该用户
	var order model.TextMessageOrder
	if err := connection.DbConnection.Where("order_no = ? AND user_id = ?", orderNo, userID).First(&order).Error; err != nil {
		logger.Logger.Errorf("Failed to find order %s for user %s: %v", orderNo, userID, err)
		err := logger.NewAppError(http.StatusNotFound, "order not found or not belong to user")
		logger.HandleError(ctx, err)
		return
	}

	if err := os.MkdirAll(replyImageSaveDir, 0o755); err != nil {
		logger.Logger.Errorf("Failed to create reply image dir: %v", err)
		err := logger.NewAppError(http.StatusInternalServerError, "failed to prepare upload directory")
		logger.HandleError(ctx, err)
		return
	}

	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), my_utils.GenerateRandomString(8), ext)
	savePath := filepath.Join(replyImageSaveDir, filename)
	if err := ctx.SaveUploadedFile(fileHeader, savePath); err != nil {
		logger.Logger.Errorf("Failed to save completion image: %v", err)
		err := logger.NewAppError(http.StatusInternalServerError, "failed to save image")
		logger.HandleError(ctx, err)
		return
	}

	imageURL := fmt.Sprintf("%s/%s", replyImageURLPrefix, filename)

	// 订单状态置为 success
	if err := connection.DbConnection.Model(&model.TextMessageOrder{}).
		Where("order_no = ?", orderNo).
		Update("status", model.OrderStatusSuccess).Error; err != nil {
		logger.Logger.Errorf("Failed to update order %s status to success: %v", orderNo, err)
		err := logger.NewAppError(http.StatusInternalServerError, "failed to update order status: "+err.Error())
		logger.HandleError(ctx, err)
		return
	}

	// 在 TextMessageDetail 写入 image_url，同时把 remarks 置为 success
	updates := map[string]interface{}{
		"image_url": imageURL,
		"remarks":   string(model.OrderStatusSuccess),
	}
	result := connection.DbConnection.Model(&model.TextMessageDetail{}).
		Where("order_no = ? AND user_id = ?", orderNo, userID).
		Updates(updates)
	if result.Error != nil {
		logger.Logger.Errorf("Failed to update TextMessageDetail image_url for order %s: %v", orderNo, result.Error)
		err := logger.NewAppError(http.StatusInternalServerError, "failed to update detail: "+result.Error.Error())
		logger.HandleError(ctx, err)
		return
	}
	if result.RowsAffected == 0 {
		logger.Logger.Warnf("No TextMessageDetail row updated for order %s (user %s)", orderNo, userID)
	}

	logger.Logger.Infof("Text message order completed: order_no=%s, image_url=%s", orderNo, imageURL)

	ctx.JSON(http.StatusOK, common.RequestResp{
		Code: http.StatusOK,
		Data: gin.H{
			"order_no":  orderNo,
			"image_url": imageURL,
			"status":    string(model.OrderStatusSuccess),
		},
		Message: "success",
	})
}

// RejectTextMessageOrder 驳回短信订单:订单状态置为 review_failed,
// 同时把驳回原因写入对应 TextMessageDetail.reject_reason
func RejectTextMessageOrder(ctx *gin.Context) {
	userID := ctx.PostForm("user_id")
	orderNo := ctx.PostForm("order_no")
	rejectReason := ctx.PostForm("reject_reason")
	if userID == "" || orderNo == "" {
		err := logger.NewAppError(http.StatusBadRequest, "user_id and order_no are required")
		logger.HandleError(ctx, err)
		return
	}
	if strings.TrimSpace(rejectReason) == "" {
		err := logger.NewAppError(http.StatusBadRequest, "reject_reason is required")
		logger.HandleError(ctx, err)
		return
	}

	var order model.TextMessageOrder
	if err := connection.DbConnection.Where("order_no = ? AND user_id = ?", orderNo, userID).First(&order).Error; err != nil {
		logger.Logger.Errorf("Failed to find order %s for user %s: %v", orderNo, userID, err)
		err := logger.NewAppError(http.StatusNotFound, "order not found or not belong to user")
		logger.HandleError(ctx, err)
		return
	}

	// 已退款的状态不再退款,避免重复退款
	var refundAmount float64
	if order.Status != model.OrderStatusFailed {
		productPriceDAO := dao.NewProductPriceDAO()
		productPrice, err := productPriceDAO.GetProductPriceByName(string(order.OrderType))
		if err != nil {
			logger.Logger.Errorf("Failed to get product price for order %s (type=%s): %v", orderNo, order.OrderType, err)
			err := logger.NewAppError(http.StatusInternalServerError, "failed to get product price: "+err.Error())
			logger.HandleError(ctx, err)
			return
		}
		refundAmount = productPrice.Price
	} else {
		logger.Logger.Infof("Order %s already in refunded state (%s), skip refund", orderNo, order.Status)
	}

	db := dao.NewDB()
	if err := db.RejectTextMessageOrder(orderNo, userID, rejectReason, refundAmount); err != nil {
		logger.Logger.Errorf("Failed to reject order %s: %v", orderNo, err)
		err := logger.NewAppError(http.StatusInternalServerError, "failed to reject order: "+err.Error())
		logger.HandleError(ctx, err)
		return
	}

	logger.Logger.Infof("Text message order rejected: order_no=%s, user_id=%s, refund=%.2f", orderNo, userID, refundAmount)

	ctx.JSON(http.StatusOK, common.RequestResp{
		Code: http.StatusOK,
		Data: gin.H{
			"order_no":      orderNo,
			"status":        string(model.OrderStatusFailed),
			"reject_reason": rejectReason,
			"refund_amount": refundAmount,
		},
		Message: "success",
	})
}
