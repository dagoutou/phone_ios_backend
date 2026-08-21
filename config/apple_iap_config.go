package config

// AppleIAPParams Apple In-App Purchase 配置
type AppleIAPParams struct {
	BundleID     string
	SharedSecret string
	Environment  string
	// App Store Server API(StoreKit 2 验证用)
	IssuerID   string // App Store Connect API Issuer ID(类似 57246542-96fe-1a63-e053-0824d011072a)
	KeyID      string // API 密钥 ID(类似 2X9R4HXF34)
	PrivateKey string // .p8 私钥内容(-----BEGIN PRIVATE KEY----- 开头)
}

const (
	IAPEnvSandbox    = "sandbox"
	IAPEnvProduction = "production"
	IAPStatusSandbox = 21007
	XiaoLu           = "xiaolu"
)

var (
	IAPVerifyURLProd = "https://buy.itunes.apple.com/verifyReceipt"
	IAPVerifyURLSand = "https://sandbox.itunes.apple.com/verifyReceipt"

	// App Store Server API(StoreKit 2)
	IAPServerAPIProd = "https://api.storekit.itunes.apple.com"
	IAPServerAPISand = "https://api.storekit-sandbox.itunes.apple.com"
)

var AppleIAPInfos = make(map[string]AppleIAPParams)

// loadAppleIAPFromSetting 在 LoadConfig 后从 yaml 加载 Apple IAP 配置到 AppleIAPInfos。
// 若 yaml 没填,使用占位符默认值,避免运行时验单 panic(会在 service 层被显式拒绝)。
func loadAppleIAPFromSetting() {
	def := AppleIAPParams{
		BundleID:     "cn.xiaolu.app",
		SharedSecret: "REPLACE_ME_SHARED_SECRET",
		Environment:  IAPEnvProduction,
	}
	if Setting != nil {
		a := Setting.APP.AppleIAP
		if a.BundleID != "" {
			def.BundleID = a.BundleID
		}
		if a.SharedSecret != "" {
			def.SharedSecret = a.SharedSecret
		}
		if a.Environment != "" {
			def.Environment = a.Environment
		}
		if a.IssuerID != "" {
			def.IssuerID = a.IssuerID
		}
		if a.KeyID != "" {
			def.KeyID = a.KeyID
		}
		if a.PrivateKey != "" {
			def.PrivateKey = a.PrivateKey
		}
	}
	AppleIAPInfos[XiaoLu] = def
}
