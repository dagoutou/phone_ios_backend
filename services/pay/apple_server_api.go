package pay

import (
	"phone_ios_backend/config"
	"phone_ios_backend/logger"
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	jwtgo "github.com/dgrijalva/jwt-go"
)

// AppleServerAPIClient App Store Server API 客户端(StoreKit 2 交易验证)
type AppleServerAPIClient struct {
	issuerID   string
	keyID      string
	privateKey string
	bundleID   string
	httpClient *http.Client
	tokenTTL   time.Duration
}

func NewAppleServerAPIClient(issuerID, keyID, privateKey, bundleID string) *AppleServerAPIClient {
	return &AppleServerAPIClient{
		issuerID:   issuerID,
		keyID:      keyID,
		privateKey: privateKey,
		bundleID:   bundleID,
		httpClient: &http.Client{Timeout: 15 * time.Second},
		tokenTTL:   20 * time.Minute,
	}
}

// Configured 是否已配置完整(缺一项都无法调 API)
func (c *AppleServerAPIClient) Configured() bool {
	return c != nil && c.issuerID != "" && c.keyID != "" &&
		strings.Contains(c.privateKey, "BEGIN PRIVATE KEY") &&
		!strings.Contains(c.issuerID, "REPLACE_ME") && !strings.Contains(c.keyID, "REPLACE_ME")
}

// ServerTransaction App Store Server API 返回的交易信息(signedTransactionInfo 的 payload)
type ServerTransaction struct {
	TransactionID         string `json:"transactionId"`
	OriginalTransactionID string `json:"originalTransactionId"`
	BundleID              string `json:"bundleId"`
	ProductID             string `json:"productId"`
	SubscriptionGroupID   string `json:"subscriptionGroupID"`
	PurchaseDate          int64  `json:"purchaseDate"`    // ms
	OriginalPurchaseDate  int64  `json:"originalPurchaseDate"`
	ExpiresDate           int64  `json:"expiresDate"`     // 订阅有效期
	Quantity              int    `json:"quantity"`
	Type                  string `json:"type"`            // Consumable/Non-Consumable/Auto-Renewable...
	AppAccountToken       string `json:"appAccountToken"`
	InAppOwnershipType    string `json:"inAppOwnershipType"`
	SignedDate            int64  `json:"signedDate"`
	Environment           string `json:"environment"`     // Sandbox/Production
	RevocationDate        int64  `json:"revocationDate"`  // >0 表示已退款/撤销
	RevocationReason      int    `json:"revocationReason"`
}

// GetTransaction 按 transactionId 查询交易。environment 决定打哪个 API。
func (c *AppleServerAPIClient) GetTransaction(ctx context.Context, transactionID, environment string) (*ServerTransaction, error) {
	if !c.Configured() {
		return nil, fmt.Errorf("app store server api not configured(need IssuerID/KeyID/PrivateKey)")
	}

	base := config.IAPServerAPIProd
	if strings.EqualFold(environment, "Sandbox") || strings.EqualFold(environment, config.IAPEnvSandbox) {
		base = config.IAPServerAPISand
	}
	url := fmt.Sprintf("%s/inApps/v1/transactions/%s", base, transactionID)

	body, err := c.getJSON(ctx, url)
	if err != nil {
		return nil, err
	}

	var resp struct {
		SignedTransactionInfo string `json:"signedTransactionInfo"`
	}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("invalid server api response: %v, body=%s", err, string(body))
	}
	if resp.SignedTransactionInfo == "" {
		return nil, fmt.Errorf("empty signedTransactionInfo, body=%s", string(body))
	}

	// 解析 JWS payload(数据来自 Apple 服务端响应,走 HTTPS,可信任)
	tx, err := decodeJWSPayload(resp.SignedTransactionInfo)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// getJSON 带 ES256 JWT 鉴权发起 GET
func (c *AppleServerAPIClient) getJSON(ctx context.Context, url string) ([]byte, error) {
	token, err := c.generateJWT()
	if err != nil {
		return nil, fmt.Errorf("generate jwt: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		// 4040010 = 交易不存在;401 = JWT 问题
		return nil, fmt.Errorf("server api http %d: %s", resp.StatusCode, truncateStr(string(raw), 300))
	}
	logger.Logger.Infof("apple server api ok: url=%s len=%d", url, len(raw))
	return raw, nil
}

// generateJWT 生成 App Store Connect API 的 ES256 JWT
func (c *AppleServerAPIClient) generateJWT() (string, error) {
	key, err := parseECPrivateKeyPEM(c.privateKey)
	if err != nil {
		return "", fmt.Errorf("parse private key: %w", err)
	}
	now := time.Now()
	token := jwtgo.NewWithClaims(jwtgo.SigningMethodES256, jwtgo.MapClaims{
		"iss": c.issuerID,
		"iat": now.Unix(),
		"exp": now.Add(c.tokenTTL).Unix(),
		"aud": "appstoreconnect-v1",
		"bid": c.bundleID,
	})
	token.Header["kid"] = c.keyID
	return token.SignedString(key)
}

// parseECPrivateKeyPEM 兼容 PKCS8(Apple .p8) 与 SEC1 两种 EC 私钥 PEM 格式
func parseECPrivateKeyPEM(pemStr string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("invalid PEM block")
	}
	if k, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if ec, ok := k.(*ecdsa.PrivateKey); ok {
			return ec, nil
		}
		return nil, errors.New("PKCS8 key is not ECDSA")
	}
	if k, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return k, nil
	}
	return nil, errors.New("unsupported key format: need EC key in PKCS8 or SEC1 PEM")
}

// decodeJWSPayload 解析 JWS 三段式结构,取出 payload 段 JSON
func decodeJWSPayload(jws string) (*ServerTransaction, error) {
	parts := strings.Split(jws, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid jws format: expect 3 segments, got %d", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		// 兼容标准 base64
		payload, err = base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, fmt.Errorf("decode jws payload: %w", err)
		}
	}
	var tx ServerTransaction
	if err = json.NewDecoder(bytes.NewReader(payload)).Decode(&tx); err != nil {
		return nil, fmt.Errorf("unmarshal transaction: %w", err)
	}
	return &tx, nil
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
