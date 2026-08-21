package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

var Setting *Config

// Config 代表整个配置文件结构
type Config struct {
	APP struct {
		Database struct {
			Host     string `yaml:"Host"`
			Port     int    `yaml:"Port"`
			User     string `yaml:"User"`
			Password string `yaml:"Password"`
			DBName   string `yaml:"DBName"`
		} `yaml:"Database"`
		Redis struct {
			Addr     string `yaml:"Addr"`
			Password string `yaml:"Password"`
			DB       int    `yaml:"DB"`
		} `yaml:"Redis"`
		Token struct {
			AiToken string `yaml:"AiToken"`
		} `yaml:"Token"`
		App struct {
			// Appid           string `yaml:"Appid"`
			// Mchid           string `yaml:"Mchid"`
			// NotifyUrl       string `yaml:"NotifyUrl"`
			AmountTotal       int    `yaml:"AmountTotal"`
			AliPayNotifyUrl   string `yaml:"AliPayNotifyUrl"`
			AliPayAmountTotal string `yaml:"AliPayAmountTotal"`
		} `yaml:"App"`
		SMS struct {
			SignName     string `yaml:"SignName"`
			TemplateCode string `yaml:"TemplateCode"`
			Count        int    `yaml:"Count"`
			Interval     int    `yaml:"Interval"`
			// 阿里云短信 AccessKey
			AccessKeyID     string `yaml:"AccessKeyID"`
			AccessKeySecret string `yaml:"AccessKeySecret"`
			// Anonymous SMS API config
			APIDomain   string `yaml:"APIDomain"`
			APIAccount  string `yaml:"APIAccount"`
			APIPassword string `yaml:"APIPassword"`
			APIProduct  string `yaml:"APIProduct"`
			SendURL     string `yaml:"SendURL"`
		} `yaml:"SMS"`
		Mail struct {
			SMTPHost    string `yaml:"smtp_host"`
			SMTPPort    int    `yaml:"smtp_port"`
			Username    string `yaml:"username"`
			Password    string `yaml:"password"`
			FromName    string `yaml:"from_name"`
			NotifyEmail string `yaml:"notify_email"`
		} `yaml:"Mail"`
		IsOnLine bool   `yaml:"IsOnLine"`
		URL1     string `yaml:"URL1"`
		URL2     string `yaml:"URL2"`
		Provider string `yaml:"Provider"`
		External struct {
			Url       string `yaml:"Url"`
			AccountNo string `yaml:"AccountNo"`
			Phone     string `yaml:"Phone"`
			Sign      string `yaml:"Sign"`
		} `yaml:"External"`
		AppleIAP struct {
			BundleID     string `yaml:"BundleID"`
			SharedSecret string `yaml:"SharedSecret"`
			Environment  string `yaml:"Environment"`
			// App Store Server API(StoreKit 2 验证)
			IssuerID   string `yaml:"IssuerID"`
			KeyID      string `yaml:"KeyID"`
			PrivateKey string `yaml:"PrivateKey"`
		} `yaml:"AppleIAP"`
	} `yaml:"APP"`
}

// LoadConfig 从配置文件中读取配置
func LoadConfig(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return err
	}
	Setting = &config
	loadAppleIAPFromSetting()
	return nil
}
