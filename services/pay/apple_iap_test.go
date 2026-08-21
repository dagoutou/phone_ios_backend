package pay

import (
	"phone_ios_backend/config"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestParseTransactionPayload 验证解析客户端传来的 transaction payload(base64 JSON / 纯 JSON)
func TestParseTransactionPayload(t *testing.T) {
	payload := map[string]interface{}{
		"transactionId":         "2000001221879465",
		"originalTransactionId": "2000001221879465",
		"bundleId":              "cn.test.app",
		"productId":             "com.test.credits",
		"purchaseDate":          1786787663000,
		"environment":           "Sandbox",
	}
	raw, _ := json.Marshal(payload)

	// base64 格式
	tx, err := parseTransactionPayload(base64.StdEncoding.EncodeToString(raw))
	if err != nil {
		t.Fatalf("base64 payload: %v", err)
	}
	if tx.TransactionID != "2000001221879465" || tx.Environment != "Sandbox" {
		t.Errorf("unexpected tx: %+v", tx)
	}

	// 纯 JSON 格式
	tx, err = parseTransactionPayload(string(raw))
	if err != nil {
		t.Fatalf("json payload: %v", err)
	}
	if tx.ProductID != "com.test.credits" {
		t.Errorf("expected product com.test.credits, got %s", tx.ProductID)
	}

	// 非法输入
	if _, err = parseTransactionPayload("not-a-payload"); err == nil {
		t.Error("expected error for invalid payload")
	}

	// 缺 transactionId
	bad, _ := json.Marshal(map[string]string{"environment": "Sandbox"})
	if _, err = parseTransactionPayload(string(bad)); err == nil {
		t.Error("expected error for missing transactionId")
	}
}

// TestNormalizeEnv 环境字段标准化
func TestNormalizeEnv(t *testing.T) {
	if got := normalizeEnv("Sandbox"); got != config.IAPEnvSandbox {
		t.Errorf("Sandbox -> %s, want %s", got, config.IAPEnvSandbox)
	}
	if got := normalizeEnv("sandbox"); got != config.IAPEnvSandbox {
		t.Errorf("sandbox -> %s, want %s", got, config.IAPEnvSandbox)
	}
	if got := normalizeEnv("Production"); got != config.IAPEnvProduction {
		t.Errorf("Production -> %s, want %s", got, config.IAPEnvProduction)
	}
}

// TestDecodeJWSPayload 验证 JWS 三段式 payload 解析
func TestDecodeJWSPayload(t *testing.T) {
	txJSON, _ := json.Marshal(ServerTransaction{
		TransactionID: "T123",
		ProductID:     "com.test.credits",
		BundleID:      "cn.test.app",
		Environment:   "Sandbox",
	})
	// header.payload.signature,base64url 编码
	jws := "eyJhbGciOiJFUzI1NiJ9." +
		base64.RawURLEncoding.EncodeToString(txJSON) +
		".c2lnbmF0dXJl"

	tx, err := decodeJWSPayload(jws)
	if err != nil {
		t.Fatalf("decode jws: %v", err)
	}
	if tx.TransactionID != "T123" || tx.ProductID != "com.test.credits" {
		t.Errorf("unexpected tx: %+v", tx)
	}

	if _, err = decodeJWSPayload("only-one-segment"); err == nil {
		t.Error("expected error for invalid jws segments")
	}
}

// TestGetTransaction_MockServer 用 mock server 模拟 App Store Server API
func TestGetTransaction_MockServer(t *testing.T) {
	txJSON, _ := json.Marshal(ServerTransaction{
		TransactionID: "T-MOCK-1",
		ProductID:     "com.test.credits",
		BundleID:      "cn.test.app",
		Environment:   "Sandbox",
	})
	jws := "eyJhbGciOiJFUzI1NiJ9." + base64.RawURLEncoding.EncodeToString(txJSON) + ".c2ln"

	var hitSandbox bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/inApps/v1/transactions/") {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		hitSandbox = true
		_ = json.NewEncoder(w).Encode(map[string]string{"signedTransactionInfo": jws})
	}))
	defer srv.Close()

	origSand := config.IAPServerAPISand
	config.IAPServerAPISand = srv.URL
	defer func() { config.IAPServerAPISand = origSand }()

	client := NewAppleServerAPIClient("issuer-id", "key-id", genTestPrivateKeyPEM(t), "cn.test.app")
	tx, err := client.GetTransaction(context.Background(), "T-MOCK-1", "Sandbox")
	if err != nil {
		t.Fatalf("get transaction: %v", err)
	}
	if !hitSandbox {
		t.Error("sandbox endpoint not hit")
	}
	if tx.TransactionID != "T-MOCK-1" {
		t.Errorf("expected T-MOCK-1, got %s", tx.TransactionID)
	}
}

// genTestPrivateKeyPEM 动态生成 P-256 私钥 PEM(mock server 不校验签名内容,但 JWT 签名需成功)
func genTestPrivateKeyPEM(t *testing.T) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}
