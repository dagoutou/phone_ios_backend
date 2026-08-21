package sms

import (
	"phone_ios_backend/config"
	"phone_ios_backend/logger"
	"phone_ios_backend/my_utils"
	"encoding/json"
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v5/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
	"math/rand"
	"strings"
	"time"
)

var SmsClient *dysmsapi20170525.Client

func CreateSMSClient() (_err error) {
	con := new(credential.Config).
		SetType("access_key").
		SetAccessKeyId(config.Setting.APP.SMS.AccessKeyID).
		SetAccessKeySecret(config.Setting.APP.SMS.AccessKeySecret)
	credential, _err := credential.NewCredential(con)
	if _err != nil {
		return _err
	}
	config := &openapi.Config{
		Credential: credential,
	}
	// Endpoint 请参考 https://api.aliyun.com/product/Dysmsapi
	config.Endpoint = tea.String("dysmsapi.aliyuncs.com")
	_result := &dysmsapi20170525.Client{}
	_result, _err = dysmsapi20170525.NewClient(config)
	SmsClient = _result
	return _err
}

func generateVerificationCode() (string, string) {
	source := rand.NewSource(time.Now().UnixNano())
	ra := rand.New(source)
	letters := []rune("1234567890")
	code := make([]rune, 4)
	for i := range code {
		code[i] = letters[ra.Intn(len(letters))]
	}
	return fmt.Sprintf("{\"code\":\"%s\"}", string(code)), string(code)
}

func SendSMSCode(phoneNumber string) (_err error) {
	code, code2 := generateVerificationCode()
	sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
		SignName:      tea.String(config.Setting.APP.SMS.SignName),
		TemplateCode:  tea.String(config.Setting.APP.SMS.TemplateCode),
		PhoneNumbers:  tea.String(phoneNumber),
		TemplateParam: tea.String(code),
	}

	runtime := &util.RuntimeOptions{}
	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		// 复制代码运行请自行打印 API 的返回值
		_, _err = SmsClient.SendSmsWithOptions(sendSmsRequest, runtime)
		if _err != nil {
			return _err
		}

		return nil
	}()
	if tryErr != nil {
		var error = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			error = _t
		} else {
			error.Message = tea.String(tryErr.Error())
		}
		// 此处仅做打印展示，请谨慎对待异常处理，在工程项目中切勿直接忽略异常。
		// 错误 message
		logger.Logger.Errorf("sms code genarate error:%v", tea.StringValue(error.Message))
		// 诊断地址
		var data interface{}
		d := json.NewDecoder(strings.NewReader(tea.StringValue(error.Data)))
		d.Decode(&data)
		if m, ok := data.(map[string]interface{}); ok {
			recommend, _ := m["Recommend"]
			logger.Logger.Errorf("Recommend %v", recommend)
		}
		_, _err = util.AssertAsString(error.Message)
		if _err != nil {
			return _err
		}
	}
	if _err = my_utils.SetPhoneCode(phoneNumber, code2, 300*time.Second); _err != nil {
		return _err
	}
	return _err
}
