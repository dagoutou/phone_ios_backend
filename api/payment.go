package api

import (
	"net/http"
	"phone_ios_backend/common"
	"phone_ios_backend/config"
	"phone_ios_backend/logger"
	phoneService "phone_ios_backend/services"
	"phone_ios_backend/services/pay"

	"github.com/gin-gonic/gin"
)

// AppleIAPVerify iOS 上传 receipt,后端调苹果验单并发放积分
func AppleIAPVerify(ctx *gin.Context) {
	var req struct {
		UserID    string `json:"user_id" binding:"required"`
		AppName   string `json:"app_name" binding:"required"`
		Receipt   string `json:"receipt" binding:"required"`
		ProductID string `json:"product_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = logger.NewAppError(http.StatusBadRequest, err.Error())
		logger.HandleError(ctx, err)
		return
	}

	// 直接从 config 获取 AppleIAPParams 创建服务
	appleIAPParams, ok := config.AppleIAPInfos[req.AppName]
	if !ok {
		logger.HandleError(ctx, logger.NewAppError(http.StatusBadRequest, "app not configured: "+req.AppName))
		return
	}
	svc := pay.NewAppleIAPService(appleIAPParams)

	result, err := svc.VerifyReceipt(ctx.Request.Context(), req.UserID, req.AppName, req.Receipt, req.ProductID)
	if err != nil {
		logger.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, common.RequestResp{
		Code:    http.StatusOK,
		Data:    result,
		Message: "success",
	})
}

// PhoneCheck 手机号在网状态校验，使用积分支付。
func PhoneCheck(ctx *gin.Context) {
	var req struct {
		UserID  string `json:"user_id" binding:"required"`
		Phone   string `json:"phone" binding:"required"`
		AppName string `json:"app_name" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.HandleError(ctx, logger.NewAppError(http.StatusBadRequest, err.Error()))
		return
	}

	// 统一完成 phone_check 商品的积分扣减和交易记录落库，不创建支付订单。
	payment, err := pay.PayWithPoints(ctx.Request.Context(), req.UserID, req.AppName, "phone_check")
	if err != nil {
		logger.HandleError(ctx, err)
		return
	}

	phoneResp, err := phoneService.PhoneCheck(req.Phone)
	if err != nil {
		logger.HandleError(ctx, err)
		return
	}
	if phoneResp.Code != 1 {
		logger.HandleError(ctx, logger.NewAppError(http.StatusInternalServerError, "手机号查询失败: "+phoneResp.Msg))
		return
	}

	ctx.JSON(http.StatusOK, common.RequestResp{
		Code: http.StatusOK,
		Data: map[string]interface{}{
			"status":         "success",
			"transaction_id": payment.TransactionID,
			"credits":        payment.Credits,
			"balance":        payment.Balance,
			"phone_check":    phoneResp.Data,
		},
		Message: "success",
	})
}
