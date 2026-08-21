package sms

import (
	"phone_ios_backend/config"
	"log"
	"testing"
)

func TestSendSMSCode(t *testing.T) {
	if err := config.LoadConfig("/Users/wanglili/study/andro/config/config.yaml"); err != nil {
		log.Fatal("loading config fatal!")
	}
	if err := CreateSMSClient(); err != nil {
		log.Fatal("get sms client fatal!")
	}
	if err := SendSMSCode("18314416271"); err != nil {
		return
	}
	t.Log("code:")
}
