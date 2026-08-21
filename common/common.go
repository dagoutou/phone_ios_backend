package common

const (
	ERROR   = "error"
	SUCCESS = "success"
)

type RequestResp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Count   int         `json:"count"`
}
