package model

import (
	"phone_ios_backend/config"
	"reflect"
	"time"
)

type PhoneOrder struct {
	ID               int       `json:"id"`
	PhoneOrderNumber string    `json:"phone_order_number"`
	PhoneCode        string    `json:"phone_code"`
	OtherPhoneNumber string    `json:"other_phone_number"`
	MyPhoneNumber    string    `json:"my_phone_number"`
	CreateTime       time.Time `json:"create_time"`
	UserID           string    `json:"user_id"`
	AppName          string    `json:"app_name"`
	OrderNo          string    `json:"order_no"`
}

func (p *PhoneOrder) TableName() string {
	return "phone_orders"
}

type PhoneNumberRequestParams struct {
	RequestUrl string `json:"requestUrl"`
	Url        string `json:"url"`
	AccountNo  string `json:"accountNo"`
	Phone      string `json:"phone"`
	Sign       string `json:"sign"`
	MyPhone    string `json:"myPhone"`
	Number     string `json:"number"` // 手机号码
	YouPhone   string `json:"youPhone"`
	PhoneCode  string `json:"phoneCode"`
	OrderNo    string `json:"orderNo"`
	Provider   string `json:"provider""`
}

func StructToMap(params PhoneNumberRequestParams) map[string]string {
	result := make(map[string]string)
	// 获取结构体的反射值
	val := reflect.ValueOf(params)
	typ := reflect.TypeOf(params)
	// 遍历结构体的字段
	for i := 0; i < val.NumField(); i++ {
		fieldValue := val.Field(i).String() // 获取字段的值
		if fieldValue != "" {               // 只添加不为空的字段
			fieldName := typ.Field(i).Tag.Get("json") // 获取json标签
			result[fieldName] = fieldValue
		}
	}
	return result
}

// func NewPhoneNumberRequestParams() *PhoneNumberRequestParams {
// 	if config.Setting.APP.IsOnLine {
// 		return &PhoneNumberRequestParams{
// 			Url:       "111",
// 			AccountNo: "127",
// 			Phone:     "16675523881",
// 			Sign:      "6AF60AC7EF33484CBF8E37D815EDAD5A",
// 		}
// 	}
// 	return &PhoneNumberRequestParams{
// 		Url:       "111",
// 		AccountNo: "133",
// 		Phone:     "13923448370",
// 		Sign:      "063C3A9BD9E742DDB00DD397B5123E88",
// 	}
// }

func NewPhoneNumberRequestParamsV2() *PhoneNumberRequestParams {
	if config.Setting.APP.IsOnLine {
		return &PhoneNumberRequestParams{
			Url:       "tao",
			AccountNo: "16675523881",
			Phone:     "16675523881",
			Sign:      "27E831486AF845ED9C84FBB41C681F6A",
		}
	}
	return &PhoneNumberRequestParams{
		Url:       "111",
		AccountNo: "133",
		Phone:     "13923448370",
		Sign:      "063C3A9BD9E742DDB00DD397B5123E88",
	}
}
