package pay

import (
	"context"
	"errors"
	"net/http"
	"phone_ios_backend/connection"
	"phone_ios_backend/dao"
	"phone_ios_backend/logger"
	"phone_ios_backend/model"
	"phone_ios_backend/my_utils"
	"time"

	"gorm.io/gorm"
)

const pointsPaymentEnvironment = "points"

// PointsPaymentResult 积分支付结果。
type PointsPaymentResult struct {
	TransactionID string `json:"transaction_id"`
	ProductName   string `json:"product_name"`
	Credits       int    `json:"credits"`
	Balance       int    `json:"balance"`
}

// PayWithPoints 统一处理积分支付：查询商品、扣减积分并记录交易。
// 业务商品从 product_prices 查询；积分套餐兼容从 iap_product_prices 查询。
// 该方法只处理支付记账，不创建 PaymentOrder 或具体业务订单。
func PayWithPoints(ctx context.Context, userID, appName, productName string) (*PointsPaymentResult, error) {
	amount, credits, err := getPointsPaymentProduct(productName)
	if err != nil {
		return nil, err
	}

	transactionID := "POINTS_" + my_utils.GenerateOrderID("PAY")
	pointsDAO := dao.NewUserPointsDAO()
	iapTxDAO := dao.NewIAPTransactionDAO()
	result := &PointsPaymentResult{
		TransactionID: transactionID,
		ProductName:   productName,
		Credits:       credits,
	}

	db := connection.DbConnection.WithContext(ctx)
	err = db.Transaction(func(tx *gorm.DB) error {
		newBalance, err := pointsDAO.Decrease(tx, userID, appName, credits,
			model.PointSourceManual, transactionID, productName, "积分支付: "+productName)
		if err != nil {
			return err
		}

		if err := iapTxDAO.CreateTx(tx, &model.IAPTransaction{
			TransactionID:         transactionID,
			OriginalTransactionID: transactionID,
			UserID:                userID,
			AppName:               appName,
			ProductID:             productName,
			Amount:                amount,
			Credits:               credits,
			Environment:           pointsPaymentEnvironment,
			ReceiptHash:           transactionID,
			PayTime:               time.Now(),
		}); err != nil {
			return err
		}

		result.Balance = newBalance
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func getPointsPaymentProduct(productName string) (float64, int, error) {
	priceDAO := dao.NewProductPriceDAO()
	productPrice, err := priceDAO.GetProductPriceByName(productName)
	if err == nil {
		credits := int(productPrice.Price)
		if credits <= 0 {
			return 0, 0, logger.NewAppError(http.StatusBadRequest, "商品积分配置无效: "+productName)
		}
		return productPrice.Price, credits, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, 0, err
	}

	iapProduct, err := dao.NewIAPProductPriceDAO().GetByName(productName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, logger.NewAppError(http.StatusNotFound, "商品不存在: "+productName)
		}
		return 0, 0, err
	}
	if iapProduct.Credits <= 0 {
		return 0, 0, logger.NewAppError(http.StatusBadRequest, "商品积分配置无效: "+productName)
	}
	return iapProduct.Price, iapProduct.Credits, nil
}
