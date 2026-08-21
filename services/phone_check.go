package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	phoneCheckURL = "https://api2.tanshuapi.com/api/phone_check/v1/index"
	phoneCheckKey = "bed6d26e818debdcb5939d8956a288b2"
)

// PhoneCheckData 对应接口返回的 data 字段
type PhoneCheckData struct {
	Phone   string `json:"phone"`
	Status  string `json:"status"`
	Isp     string `json:"isp"`
	IsXhzw  string `json:"is_xhzw"`
	IspReal string `json:"isp_real"`
}

// PhoneCheckResp 接口完整响应
type PhoneCheckResp struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	Data PhoneCheckData `json:"data"`
}

// PhoneCheck 直连运营商权威核验手机号在网状态
func PhoneCheck(phone string) (PhoneCheckResp, error) {
	var resp PhoneCheckResp

	if phone == "" {
		return resp, fmt.Errorf("手机号不能为空")
	}

	params := url.Values{}
	params.Set("key", phoneCheckKey)
	params.Set("phone", phone)

	req, err := http.NewRequest("GET", phoneCheckURL+"?"+params.Encode(), nil)
	if err != nil {
		return resp, err
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return resp, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return resp, err
	}

	if err = json.Unmarshal(body, &resp); err != nil {
		return resp, fmt.Errorf("JSON 解析失败: %v, body: %s", err, string(body))
	}
	return resp, nil
}
