package pay

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"phone_ios_backend/config"
	"phone_ios_backend/connection"
	"phone_ios_backend/dao"
	"phone_ios_backend/logger"
	"phone_ios_backend/model"
	"strings"
	"time"

	"gorm.io/gorm"
)

// AppleIAPService 苹果内购验单 + 积分发放服务(StoreKit 2)
type AppleIAPService struct {
	bundleID  string
	serverAPI *AppleServerAPIClient
	db        *gorm.DB
}

func NewAppleIAPService(params config.AppleIAPParams) *AppleIAPService {
	return &AppleIAPService{
		bundleID:  params.BundleID,
		serverAPI: NewAppleServerAPIClient(params.IssuerID, params.KeyID, params.PrivateKey, params.BundleID),
		db:        connection.DbConnection,
	}
}

// VerifyResult 验单成功返回给 APP 的数据
type VerifyResult struct {
	TransactionID string `json:"transaction_id"`
	ProductID     string `json:"product_id"`
	Credits       int    `json:"credits"`
	Balance       int    `json:"balance"`
	Environment   string `json:"environment"`
}

// VerifyReceipt 主入口:验证 StoreKit 2 交易并发放积分。
// receiptPayload 为客户端 uni.requestVirtualPayment 返回的 transaction.jsonRepresentation(base64 JSON)。
func (s *AppleIAPService) VerifyReceipt(ctx context.Context, userID, appName, receiptPayload, productID string) (*VerifyResult, error) {
	if s == nil {
		return nil, errors.New("apple iap service not initialized")
	}
	if userID == "" || receiptPayload == "" || productID == "" {
		return nil, errors.New("missing required params")
	}
	if !s.serverAPI.Configured() {
		return nil, errors.New("app store server api not configured: need IssuerID/KeyID/PrivateKey in config.yaml")
	}

	// 1. receipt hash + redis 锁,防并发重复提交
	receiptHash := sha256Hex(receiptPayload)
	lockKey := "iap:verify:" + receiptHash
	locked, err := connection.Rdb.SetNX(ctx, lockKey, "1", 30*time.Second).Result()
	if err == nil && !locked {
		return nil, logger.NewAppError(http.StatusTooManyRequests, "receipt is being processed")
	}
	defer func() {
		if err == nil && locked {
			_ = connection.Rdb.Del(ctx, lockKey).Err()
		}
	}()

	// 2. 解析客户端传的 transaction payload,提取 transactionId/environment
	clientTx, err := parseTransactionPayload(receiptPayload)
	if err != nil {
		return nil, logger.NewAppError(http.StatusBadRequest, "invalid receipt payload: "+err.Error())
	}
	if clientTx.BundleID != s.bundleID {
		return nil, logger.NewAppError(http.StatusBadRequest,
			fmt.Sprintf("bundle_id mismatch: got=%s want=%s", clientTx.BundleID, s.bundleID))
	}
	if clientTx.ProductID != productID {
		return nil, logger.NewAppError(http.StatusBadRequest,
			fmt.Sprintf("product_id mismatch: payload=%s request=%s", clientTx.ProductID, productID))
	}

	// 3. 调 App Store Server API 按 transactionId 查询真实交易
	serverTx, err := s.serverAPI.GetTransaction(ctx, clientTx.TransactionID, clientTx.Environment)
	if err != nil {
		return nil, logger.NewAppError(http.StatusBadRequest, "apple server verify failed: "+err.Error())
	}
	// 4. 校验服务端返回与客户端声明一致,且未被退款撤销
	if serverTx.TransactionID != clientTx.TransactionID {
		return nil, logger.NewAppError(http.StatusBadRequest, "transaction_id mismatch after server verify")
	}
	if serverTx.ProductID != productID {
		return nil, logger.NewAppError(http.StatusBadRequest,
			fmt.Sprintf("product_id mismatch after server verify: got=%s want=%s", serverTx.ProductID, productID))
	}
	if serverTx.BundleID != s.bundleID {
		return nil, logger.NewAppError(http.StatusBadRequest, "bundle_id mismatch after server verify")
	}
	if serverTx.RevocationDate > 0 {
		return nil, logger.NewAppError(http.StatusBadRequest, "transaction was revoked/refunded")
	}

	// 5. 防重放:同 transaction_id 不能重复发放
	iapDAO := dao.NewIAPTransactionDAO()
	exists, err := iapDAO.ExistsByTransactionID(serverTx.TransactionID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, logger.NewAppError(http.StatusConflict, "transaction already granted")
	}

	// 6. 查 iap_product_prices 拿 credits/amount
	iapPriceDAO := dao.NewIAPProductPriceDAO()
	pp, err := iapPriceDAO.GetByProductID(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, logger.NewAppError(http.StatusBadRequest, "product not configured: "+productID)
		}
		return nil, err
	}
	if pp.Credits <= 0 {
		return nil, logger.NewAppError(http.StatusInternalServerError, "product credits not configured: "+productID)
	}

	// 7. 在事务里:写 IAP 交易 + 加积分
	usedEnv := normalizeEnv(serverTx.Environment)
	payTime := parseMsTime(serverTx.PurchaseDate)
	pointsDAO := dao.NewUserPointsDAO()
	var newBalance int
	err = s.db.Transaction(func(tx *gorm.DB) error {
		record := &model.IAPTransaction{
			TransactionID:         serverTx.TransactionID,
			OriginalTransactionID: serverTx.OriginalTransactionID,
			UserID:                userID,
			AppName:               appName,
			ProductID:             productID,
			Amount:                pp.Price,
			Credits:               pp.Credits,
			Environment:           usedEnv,
			ReceiptHash:           receiptHash,
			PayTime:               payTime,
		}
		if err := iapDAO.CreateTx(tx, record); err != nil {
			return err
		}
		// phone 类型商品:同步创建通话套餐订单记录
		if strings.HasPrefix(pp.Name, "phone") {
			phoneOrder := &model.PhoneOrder{
				PhoneOrderNumber: "IAP" + serverTx.TransactionID,
				PhoneCode:        pp.Name,
				OrderNo:          serverTx.TransactionID,
				UserID:           userID,
				AppName:          appName,
				CreateTime:       time.Now(),
			}
			if err := tx.Create(phoneOrder).Error; err != nil {
				return err
			}
		}
		remark := "Apple IAP 充值"
		if usedEnv == config.IAPEnvSandbox {
			remark = "Apple IAP 沙盒测试充值"
		}
		bal, err := pointsDAO.Increase(tx, userID, appName, pp.Credits,
			model.PointSourceIAP, serverTx.OriginalTransactionID, productID, remark)
		if err != nil {
			return err
		}
		newBalance = bal
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &VerifyResult{
		TransactionID: serverTx.TransactionID,
		ProductID:     productID,
		Credits:       pp.Credits,
		Balance:       newBalance,
		Environment:   usedEnv,
	}, nil
}

// parseTransactionPayload 解析客户端传来的交易数据。
// 支持 base64(JSON) 与纯 JSON 两种格式。
func parseTransactionPayload(payload string) (*ServerTransaction, error) {
	raw := payload
	if decoded, err := base64.StdEncoding.DecodeString(payload); err == nil {
		raw = string(decoded)
	}
	var tx ServerTransaction
	if err := json.Unmarshal([]byte(raw), &tx); err != nil {
		return nil, fmt.Errorf("not valid json/base64-json: %v", err)
	}
	if tx.TransactionID == "" {
		return nil, errors.New("transactionId missing in payload")
	}
	if tx.Environment == "" {
		return nil, errors.New("environment missing in payload")
	}
	return &tx, nil
}

// normalizeEnv Sandbox→sandbox,Production→production
func normalizeEnv(env string) string {
	switch env {
	case "Sandbox", config.IAPEnvSandbox:
		return config.IAPEnvSandbox
	default:
		return config.IAPEnvProduction
	}
}

func parseMsTime(ms int64) time.Time {
	if ms <= 0 {
		return time.Now()
	}
	return time.UnixMilli(ms)
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
