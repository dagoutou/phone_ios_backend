package my_utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	JwtSecret      string = "AWGE123ifaswwf423nsaf232s1se" // JWT 密钥
	ALIPAY_GATEWAY        = "https://openapi.alipay.com/gateway.do"
)

type Claims struct {
	OpenID string `json:"openid"`
	jwt.StandardClaims
}

type PhoneNumberResponse struct {
	ErrCode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
	PhoneInfo struct {
		PhoneNumber     string `json:"phoneNumber"`
		PurePhoneNumber string `json:"purePhoneNumber"`
		CountryCode     string `json:"countryCode"`
		AccessToken     string `json:"accessToken"`
	} `json:"phone_info"`
}

func GenerateJWT(openid string) (string, error) {
	claims := Claims{
		OpenID: openid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 15 * 180).Unix(), // 15天有效期
			Issuer:    "your_app",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JwtSecret))
}

// ParseToken 解析 Token 并验证
func ParseToken(tokenString string) (*Claims, error) {
	// 解析 JWT
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证 token 的签名方法是否正确
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}
		return []byte(JwtSecret), nil // 返回密钥用于验证签名
	})

	if err != nil {
		if err.Error() == "token is expired" {
			return nil, errors.New("token 已过期")
		}
		return nil, err
	}

	// 获取 claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的 token")
}
