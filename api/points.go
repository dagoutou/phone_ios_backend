package api

import (
	"phone_ios_backend/common"
	"phone_ios_backend/connection"
	"phone_ios_backend/dao"
	"phone_ios_backend/logger"
	"phone_ios_backend/model"
	"phone_ios_backend/my_utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
		Code:  http.StatusOK,
		Data:  map[string]interface{}{"list": list, "total": total},
		Count: int(total),
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
		UserID     string `json:"user_id" binding:"required"`
		AppName    string `json:"app_name" binding:"required"`
		ProductName string `json:"product_name" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = logger.NewAppError(http.StatusBadRequest, err.Error())
		logger.HandleError(ctx, err)
		return
	}

	// 生成唯一交易ID
	transactionID := "POINTS_" + my_utils.GenerateOrderID("PAY")

	// 从 product_prices 表查询商品价格
	priceDAO := dao.NewProductPriceDAO()
	productPrice, err := priceDAO.GetProductPriceByName(req.ProductName)
	if err != nil {
		logger.HandleError(ctx, logger.NewAppError(http.StatusNotFound, "商品不存在: "+err.Error()))
		return
	}

	// 商品价格作为积分扣减数量(1元=1积分)
	credits := int(productPrice.Price)

	pointsDAO := dao.NewUserPointsDAO()
	iapTxDAO := dao.NewIAPTransactionDAO()

	var result struct {
		TransactionID string `json:"transaction_id"`
		ProductName   string `json:"product_name"`
		Credits       int    `json:"credits"`
		Balance       int    `json:"balance"`
	}

	// 在事务中完成：扣减积分 + 记录交易
	err = connection.DbConnection.Transaction(func(tx *gorm.DB) error {
		// 1. 扣减积分
		newBalance, err := pointsDAO.Decrease(tx, req.UserID, req.AppName, credits,
			model.PointSourceManual, transactionID, req.ProductName, "积分支付: "+req.ProductName)
		if err != nil {
			return err
		}

		// 2. 记录交易信息到 IAPTransaction
		iapTx := &model.IAPTransaction{
			TransactionID:         transactionID,
			OriginalTransactionID: transactionID,
			UserID:                req.UserID,
			AppName:               req.AppName,
			ProductID:             req.ProductName,
			Amount:                productPrice.Price,
			Credits:               credits,
			Environment:           "points",
			ReceiptHash:           transactionID,
			PayTime:               time.Now(),
		}
		if err := iapTxDAO.CreateTx(tx, iapTx); err != nil {
			return err
		}

		result.TransactionID = transactionID
		result.ProductName = req.ProductName
		result.Credits = credits
		result.Balance = newBalance
		return nil
	})

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
