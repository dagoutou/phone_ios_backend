package sms

import (
	"phone_ios_backend/config"
	"phone_ios_backend/logger"
	"phone_ios_backend/model"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SMSService 短信服务
type SMSService struct {
	client *http.Client
	config *config.Config
}

// NewSMSService 创建短信服务
func NewSMSService() *SMSService {
	return &SMSService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: config.Setting,
	}
}

// SendSMSRequest 短信发送请求
type SendSMSRequest struct {
	APIDomain    string `json:"api_domain"`
	APIAccount   string `json:"api_account"`
	APIPassword  string `json:"api_password"`
	APIProduct   string `json:"api_product"`
	Content      string `json:"content"`
	Houzhui      string `json:"houzhui,omitempty"`
	Mobile       string `json:"mobile"`
	Account      string `json:"account"`
	PlanTime     *int64 `json:"plan_time,omitempty"`
	SMSAutograph string `json:"sms_autograph"`
}

// SendSMSResponse 短信发送响应
type SendSMSResponse struct {
	Bool bool     `json:"bool"`
	Msg  string   `json:"msg"`
	Type string   `json:"type"`
	Data *SMSData `json:"data,omitempty"`
}

// SMSData 短信发送数据
type SMSData struct {
	Tel          string `json:"tel"`
	Domain       string `json:"domain"`
	Mobile       string `json:"mobile"`
	Account      string `json:"account"`
	Content      string `json:"content"`
	PlanTime     int64  `json:"plan_time"`
	CreateTime   int64  `json:"create_time"`
	SMSAutograph string `json:"sms_autograph"`
	APISendID    string `json:"api_send_id"`
	Strip        int    `json:"strip"`
	Remarks      string `json:"remarks"`
}

// SendSMSParams 发送短信参数
type SendSMSParams struct {
	Content      string // 短信内容
	Mobile       string // 收件人号码
	Account      string // 发件人账号
	PlanTime     *int64 // 定时发送时间（秒级时间戳），不填默认立即发送
	SMSAutograph string // 短信签名ID
	Houzhui      string // 短信后缀（可选）
}

// SendTextMessage 发送短信
func (s *SMSService) SendTextMessage(params SendSMSParams) (*SendSMSResponse, error) {
	// 构建请求
	req := SendSMSRequest{
		APIDomain:    s.config.APP.SMS.APIDomain,
		APIAccount:   s.config.APP.SMS.APIAccount,
		APIPassword:  s.config.APP.SMS.APIPassword,
		APIProduct:   s.config.APP.SMS.APIProduct,
		Content:      params.Content,
		Mobile:       params.Mobile,
		Account:      params.Account,
		PlanTime:     params.PlanTime,
		SMSAutograph: params.SMSAutograph,
		Houzhui:      params.Houzhui,
	}

	// 序列化请求
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	// 发送请求
	url := s.config.APP.SMS.SendURL
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	logger.Logger.Infof("Sending SMS request: %s", string(jsonData))

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var smsResp SendSMSResponse
	if err := json.NewDecoder(resp.Body).Decode(&smsResp); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	logger.Logger.Infof("SMS response: bool=%v, msg=%s", smsResp.Bool, smsResp.Msg)

	return &smsResp, nil
}

// ResponseToDetail 将API响应转换为TextMessageDetail
func ResponseToDetail(resp *SendSMSResponse, userID, orderNo, content string, msgType model.MessageDetailType) *model.TextMessageDetail {
	if resp.Data == nil {
		return nil
	}

	detail := &model.TextMessageDetail{
		APISendID: resp.Data.APISendID,
		UserID:    userID,
		OrderNo:   orderNo,
		Mobile:    resp.Data.Mobile,
		Account:   resp.Data.Account,
		Content:   content,
		Strip:     resp.Data.Strip,
		Remarks:   "success",
		Domain:    resp.Data.Domain,
		Type:      msgType,
		CreatedAt: time.Now(),
	}

	// 转换时间戳
	if resp.Data.PlanTime > 0 {
		t := time.Unix(resp.Data.PlanTime, 0)
		detail.PlanTime = &t
	}

	if resp.Data.CreateTime > 0 {
		t := time.Unix(resp.Data.CreateTime, 0)
		detail.CreateTime = &t
	}

	return detail
}
