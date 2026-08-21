package connection

import (
	"fmt"
	"phone_ios_backend/config"
	"phone_ios_backend/model"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DbConnection *gorm.DB
	Rdb          *redis.Client
)

// ConnectDatabase 使用配置连接数据库
func ConnectDatabase(config *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.APP.Database.User,
		config.APP.Database.Password,
		config.APP.Database.Host,
		config.APP.Database.Port,
		config.APP.Database.DBName,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)
	DbConnection = db
	if err = db.AutoMigrate(&model.PhoneOrder{}, &model.BlackNumber{}, &model.UserRealNameAuthentication{},
		&model.PhonePlan{}, &model.User{}, &model.PriceMenu{},
		&model.TextMessageOrder{}, &model.TextMessageDetail{}, &model.TextMessageReply{}, &model.ProductPrice{},
		&model.IAPProductPrice{},
		&model.IAPTransaction{}, &model.UserPoints{}, &model.PointTransaction{}); err != nil {
		return nil, fmt.Errorf("auto migrate failed: %w", err)
	}
	if err = initData(db); err != nil {
		return nil, err
	}
	return db, nil
}

func initData(db *gorm.DB) error {
	var count int64
	// 检查表里是否已经有数据
	if err := db.Model(&model.PhonePlan{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		// 插入初始数据
		plans := []model.PhonePlan{
			{PhoneCode: "90000", PhoneName: "一个国内隐私号，XXXXX描述！"},
			{PhoneCode: "70000", PhoneName: "一个国内临时号，XXXXX描述！"},
		}
		if err := db.Create(&plans).Error; err != nil {
			return err
		}
	}

	// 检查表里是否已经有数据
	if err := db.Model(&model.PriceMenu{}).Count(&count).Error; err != nil {
		return err
	}
	plans := []model.PriceMenu{
		{
			ID:       1,
			Menu:     "基础体验套餐",
			Price:    20,
			Amount:   20,
			Describe: "到账总额 20 元。适合新用户体验，日常使用更轻松！",
		},
		{
			ID:       2,
			Menu:     "热门推荐套餐",
			Price:    40,
			Amount:   45,
			Describe: "充值 40 元到账 45 元，立省 5 元。兼顾实用与优惠，性价比之选！",
		},
		{
			ID:       3,
			Menu:     "超值畅享套餐",
			Price:    80,
			Amount:   90,
			Describe: "支付 80 元，到账 90 元，立赠 10 元！通话无烦恼，为深度用户量身打造，尽享高额回馈！",
		},
	}
	if count == 0 {
		if err := db.Create(&plans).Error; err != nil {
			return err
		}
	} else {
		for _, p := range plans {
			if err := db.Model(&model.PriceMenu{}).
				Where("id = ?", p.ID).
				Updates(map[string]interface{}{
					"menu":     p.Menu,
					"price":    p.Price,
					"amount":   p.Amount,
					"describe": p.Describe,
				}).Error; err != nil {
				return err
			}
		}
	}
	// 初始化或更新商品价格数据
	if err := db.Model(&model.ProductPrice{}).Count(&count).Error; err != nil {
		return err
	}
	productPrices := []model.ProductPrice{
		{
			Name:  "text_message",
			Price: 3,
		},
		{
			Name:  "manual_message",
			Price: 9.9,
		},
		{
			Name:  "phone_check",
			Price: 3,
		},
	}
	if count == 0 {
		if err := db.Create(&productPrices).Error; err != nil {
			return err
		}
	} else {
		for _, p := range productPrices {
			if err := db.Model(&model.ProductPrice{}).
				Where("name = ?", p.Name).
				Update("price", p.Price).Error; err != nil {
				return err
			}
		}
	}
	// 初始化或更新 Apple IAP 内购套餐
	iapProducts := []model.IAPProductPrice{
		{Name: "credits6", ProductID: "com.lovephone.1", Price: 6, Credits: 6},
		{Name: "credits30", ProductID: "com.lovephone.2", Price: 30, Credits: 30},
		{Name: "credits68", ProductID: "com.lovephone.3", Price: 68, Credits: 68},
	}
	for _, p := range iapProducts {
		if err := db.Where(model.IAPProductPrice{ProductID: p.ProductID}).
			Assign(map[string]interface{}{
				"name":    p.Name,
				"price":   p.Price,
				"credits": p.Credits,
			}).
			FirstOrCreate(&p).Error; err != nil {
			return err
		}
	}
	return nil
}

func InitRedis() {
	Rdb = redis.NewClient(&redis.Options{
		Addr:         config.Setting.APP.Redis.Addr,
		Password:     config.Setting.APP.Redis.Password,
		DB:           config.Setting.APP.Redis.DB,
		PoolSize:     10,              // 连接池最大连接数（默认是 CPU * 10）
		MinIdleConns: 5,               // 最小空闲连接数
		DialTimeout:  5 * time.Second, // 连接超时时间
		ReadTimeout:  3 * time.Second, // 读超时
		WriteTimeout: 3 * time.Second, // 写超时
		PoolTimeout:  4 * time.Second, // 从连接池拿连接的最大等待时间
	})
}
