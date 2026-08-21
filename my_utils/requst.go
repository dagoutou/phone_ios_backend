package my_utils

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"phone_ios_backend/connection"
	"phone_ios_backend/logger"
	"phone_ios_backend/model"
	"sort"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// 计算 MD5 哈希
func md5Encrypt(text string) string {
	hash := md5.Sum([]byte(text))                       // 计算 MD5
	return strings.ToUpper(hex.EncodeToString(hash[:])) // 返回 16 进制字符串
}

// 对请求参数进行 ASCII 排序并计算 MD5 签名
func generateSignature(params map[string]string) string {
	// 获取所有键并进行 ASCII 升序排序
	keys := make([]string, 0, len(params))
	for k := range params {
		// 排除 sign 和 requestUrl
		if k != "sign" && k != "requestUrl" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	// 生成 key=value 形式的字符串
	var paramStr []string
	for _, k := range keys {
		paramStr = append(paramStr, fmt.Sprintf("%s=%s", k, params[k]))
	}
	// 拼接字符串后计算 MD5
	signStr := strings.Join(paramStr, "&")
	// 将 sign 字段放到最后
	if sign, exists := params["sign"]; exists {
		signStr += sign
	}
	return md5Encrypt(signStr)
}

// SendPostRequest 发送 POST 请求，返回 JSON 格式的数据
func SendPostRequest(apiURL string, params *model.PhoneNumberRequestParams) (map[string]interface{}, error) {
	mp := model.StructToMap(*params)
	// 计算签名
	signature := generateSignature(mp)
	var data = make(map[string]string)
	data["sign"] = signature
	for k, v := range mp {
		if k == "requestUrl" || k == "sign" {
			continue
		}
		data[k] = v
	}
	ms, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(ms))
	if err != nil {
		return nil, err
	}
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送 HTTP 请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// 解析 JSON 响应
	var jsonResponse map[string]interface{}
	if err := json.Unmarshal(body, &jsonResponse); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %v", err)
	}
	return jsonResponse, nil
}

const SuccessCode = "0000"

func GetRealNameAuthCount(userID string) (int64, error) {
	return decreaseValue(userID)
}

func decreaseValue(userID string) (int64, error) {
	key := fmt.Sprintf("app:%s", userID)
	// 使用 Lua 脚本保证原子性（防止并发问题）
	script := redis.NewScript(`
		local ttl = redis.call("TTL", KEYS[1])
		if ttl < 0 then
			redis.call("SET", KEYS[1], 10, "EX", 86400)
			return 5
		end
	
		local current = tonumber(redis.call("GET", KEYS[1]) or "0")
		if current <= 0 then
			return 0
		else
			current = current - 1
			redis.call("SET", KEYS[1], current, "KEEPTTL")
			return current
		end
	`)
	ctx := context.Background()
	// 执行脚本
	result, err := script.Run(ctx, connection.Rdb, []string{key}).Result()
	if err != nil {
		return 0, err
	}

	value, ok := result.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected result type")
	}
	str := "===" + key + ":剩余%d次"
	logger.Logger.Infof(str, value)
	return value, nil
}

type AliyunRealNameResp struct {
	Code       int    `json:"code"`
	Desc       string `json:"desc"`
	Similarity int    `json:"similarity"`
	Data       struct {
		Birthday string `json:"birthday"`
		Address  string `json:"address"`
		Sex      string `json:"sex"`
	} `json:"data"`
}

func AliyunRealName(userName, idCard, faceImage string) (AliyunRealNameResp, error) {
	var resp AliyunRealNameResp
	host := "https://vidface.market.alicloudapi.com"
	path := "/lundear/idface"
	appCode := "a3ff2dd178ce4386859ba20201d6eb11" // 请替换为你自己的AppCode

	// 构建 URL
	endpoint := host + path

	// 构建请求 body（form 格式）
	form := url.Values{}
	form.Set("idcard", idCard)
	form.Set("name", userName)
	form.Set("image", faceImage)
	form.Set("liveck", "1")
	body := form.Encode()

	// 构建请求
	req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(body))
	if err != nil {
		panic(err)
	}

	// 设置 Headers
	req.Header.Set("Authorization", "APPCODE "+appCode)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	// 发送请求
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return resp, err
	}
	if err = json.Unmarshal(data, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// SetPhoneCode 设置验证码：key 为 phone_number，value 为 code，带过期时间
func SetPhoneCode(phone string, code string, ttl time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf("verify_code:%s", phone)
	return connection.Rdb.Set(ctx, key, code, ttl).Err()
}

// GetPhoneCode 获取验证码
func GetPhoneCode(phone string) (string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("verify_code:%s", phone)
	val, err := connection.Rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		// Key 不存在，返回空字符串或你自定义的错误
		return "", nil // 或 return "", errors.New("验证码不存在")
	} else if err != nil {
		// 其他错误，例如连接失败等
		return "", err
	}
	return val, nil
}

// 获取当日 23:59:59 剩余秒数
func getSecondsUntilEndOfDay() time.Duration {
	now := time.Now()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	return endOfDay.Sub(now)
}

// CanSendCode 检查并增加发送次数
func CanSendCode(phone string) (bool, error) {
	ctx := context.Background()
	date := time.Now().Format("2006-01-02") // e.g. "2025-05-06"
	key := fmt.Sprintf("send_count:%s:%s", phone, date)

	count, err := connection.Rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		// 第一次发送，设置当天过期时间
		connection.Rdb.Expire(ctx, key, getSecondsUntilEndOfDay())
	}

	if count > 10 {
		return false, nil
	}
	return true, nil
}

func GenerateOrderID(appType string) string {
	// 获取当前时间
	now := time.Now()
	dateStr := now.Format("20060102150405")
	// 生成 6 位随机数
	so := rand.NewSource(time.Now().UnixNano())
	ran := rand.New(so)
	randomNum := ran.Intn(1000000)
	randomStr := fmt.Sprintf("%06d", randomNum)
	// 组合订单号 (时间戳 + 随机数)
	orderID := fmt.Sprintf("%s%s%s", appType, dateStr, randomStr)
	return orderID
}

func FixTsIfBeforeNow(ts int64) int64 {
	if ts < time.Now().Unix() {
		return ts + int64(8*time.Hour/time.Second)
	}
	return ts
}

// SendPostRequestUrl 发送 POST 请求，返回 JSON 格式的数据
func SendPostRequestUrl(apiURL string, params *model.PhoneNumberRequestParams) ([]byte, error) {
	mp := model.StructToMap(*params)
	// 计算签名
	signature := generateSignature(mp)
	var data = make(map[string]string)
	data["sign"] = signature
	for k, v := range mp {
		if k == "requestUrl" || k == "sign" {
			continue
		}
		data[k] = v
	}
	ms, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(ms))
	if err != nil {
		return nil, err
	}
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送 HTTP 请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// 解析 JSON 响应
	return body, nil
}
