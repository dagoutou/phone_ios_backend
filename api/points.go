package api

import (
	"net/http"
	"phone_ios_backend/common"
	"phone_ios_backend/dao"
	"phone_ios_backend/logger"
	"phone_ios_backend/services/pay"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetPointsBalance 查询用户积分余额
func GetPointsBalance(ctx *gin.Context) {
	userID := ctx.Query("user_id")
	appName := ctx.Query("app_name")
	if userID == "" || appName == "" {
		ctx.JSON(http.StatusOK, common.RequestResp{
			Code:    http.StatusBadRequest,
			Message: "missing user_id or app_name",
		})
		return
	}
	pointsDAO := dao.NewUserPointsDAO()
	p, err := pointsDAO.GetByUser(userID, appName)
	if err != nil {
		logger.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, common.RequestResp{
		Code: http.StatusOK,
		Data: map[string]interface{}{
			"user_id": userID,
			"balance": p.Balance,
		},
		Message: "success",
	})
}

// GetPointTransactions 分页查询用户积分明细
func GetPointTransactions(ctx *gin.Context) {
	userID := ctx.Query("user_id")
	appName := ctx.Query("app_name")
	if userID == "" || appName == "" {
		ctx.JSON(http.StatusOK, common.RequestResp{
			Code:    http.StatusBadRequest,
			Message: "missing user_id or app_name",
		})
		return
	}
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(ctx.DefaultQuery("size", "20"))
	pointsDAO := dao.NewUserPointsDAO()
	list, total, err := pointsDAO.ListTransactionsByUser(userID, appName, page, size)
	if err != nil {
		logger.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, common.RequestResp{
		Code:    http.StatusOK,
		Data:    map[string]interface{}{"list": list, "total": total},
		Count:   int(total),
		Message: "success",
	})
}

// GetAppleIAPProducts 获取 Apple IAP 商品列表(返回完整信息)
func GetAppleIAPProducts(ctx *gin.Context) {
	iapProductDAO := dao.NewIAPProductPriceDAO()
	products, err := iapProductDAO.ListAll()
	if err != nil {
		logger.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.RequestResp{
		Code:    http.StatusOK,
		Data:    products,
		Message: "success",
	})
}

// PointsPayment 使用积分支付商品
func PointsPayment(ctx *gin.Context) {
	var req struct {
		UserID      string `json:"user_id" binding:"required"`
		AppName     string `json:"app_name" binding:"required"`
		ProductName string `json:"product_name" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = logger.NewAppError(http.StatusBadRequest, err.Error())
		logger.HandleError(ctx, err)
		return
	}

	result, err := pay.PayWithPoints(ctx.Request.Context(), req.UserID, req.AppName, req.ProductName)
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
