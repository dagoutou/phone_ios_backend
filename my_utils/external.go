package my_utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"phone_ios_backend/config"
	"phone_ios_backend/logger"
)

// ExternalResp 与上游 /external/* 接口返回结构保持一致
type ExternalResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

var externalClient = &http.Client{
	Timeout: 10 * time.Second,
}

func externalUpstreamURL(path string) string {
	return config.Setting.APP.External.Url + path
}

func applyExternalHeaders(req *http.Request) {
	ext := config.Setting.APP.External
	req.Header.Set("X-Account-No", ext.AccountNo)
	req.Header.Set("X-Phone", ext.Phone)
	req.Header.Set("X-Sign", ext.Sign)
}

func parseExternalResp(resp *http.Response) (ExternalResp, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Logger.Errorf("external read response failed: %v", err)
		return ExternalResp{}, err
	}
	var result ExternalResp
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Logger.Errorf("external parse response failed: %v, body=%s", err, string(body))
		return ExternalResp{}, fmt.Errorf("parse response failed: %w", err)
	}
	return result, nil
}

// SendExternalGet 调用上游 /external/* GET 接口
// 返回解析后的响应、HTTP 状态码以及传输/解析错误
func SendExternalGet(path string, query url.Values) (ExternalResp, int, error) {
	req, err := http.NewRequest("GET", externalUpstreamURL(path), nil)
	if err != nil {
		logger.Logger.Errorf("external GET create request failed: %v", err)
		return ExternalResp{}, 0, err
	}
	if query != nil {
		req.URL.RawQuery = query.Encode()
	}
	applyExternalHeaders(req)

	resp, err := externalClient.Do(req)
	if err != nil {
		logger.Logger.Errorf("external GET upstream failed: %v", err)
		return ExternalResp{}, http.StatusBadGateway, err
	}
	defer resp.Body.Close()

	result, perr := parseExternalResp(resp)
	return result, resp.StatusCode, perr
}

// SendExternalPost 调用上游 /external/* POST 接口
// 返回解析后的响应、HTTP 状态码以及传输/解析错误
func SendExternalPost(path string, body []byte) (ExternalResp, int, error) {
	req, err := http.NewRequest("POST", externalUpstreamURL(path), bytes.NewReader(body))
	if err != nil {
		logger.Logger.Errorf("external POST create request failed: %v", err)
		return ExternalResp{}, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	applyExternalHeaders(req)

	resp, err := externalClient.Do(req)
	if err != nil {
		logger.Logger.Errorf("external POST upstream failed: %v", err)
		return ExternalResp{}, http.StatusBadGateway, err
	}
	defer resp.Body.Close()

	result, perr := parseExternalResp(resp)
	return result, resp.StatusCode, perr
}
