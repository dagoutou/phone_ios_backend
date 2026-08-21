package api

import (
	"net/http"
	"phone_ios_backend/common"
	"phone_ios_backend/config"
	"phone_ios_backend/logger"
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
