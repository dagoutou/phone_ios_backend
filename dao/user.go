package dao

import (
	"gorm.io/gorm"

	"phone_ios_backend/model"
	"time"
)

func (d *DBConnection) CreateUser(phoneNumber string) (model.User, error) {
	user := model.User{
		UserID:     phoneNumber,
		CreateTime: time.Now(),
	}
	var err error
	if err = d.db.Create(&user).Error; err != nil {
		return model.User{}, err
	}
	// 线上模式新用户赠送 3 元余额；非线上模式（测试/演示）赠 40 元
	// var balance float64
	// if config.Setting.APP.IsOnLine {
	// 	balance = 3
	// } else {
	// 	balance = 40
	// }
	// ba := model.UserBalance{
	// 	UserID:     user.UserID,
	// 	Balance:    balance,
	// 	CreateTime: time.Now(),
	// }
	// if err = d.db.Create(&ba).Error; err != nil {
	// 	return model.User{}, err
	// }
	return user, err
}

func (d *DBConnection) GetUserByID(userID string) (bool, model.User, error) {
	var (
		m   model.User
		cnt int64
	)
	err := d.db.Table("users").Where("user_id=?", userID).Count(&cnt).First(&m).Error
	return cnt > 0, m, err
}

type UserBalance struct {
	UserID  string  `json:"user_id"`
	Balance float64 `json:"balance"` // 余额
}

func (d *DBConnection) GetPriceMenu() ([]model.PriceMenu, error) {
	var menus []model.PriceMenu
	err := d.db.Model(&model.PriceMenu{}).Find(&menus).Error
	return menus, err
}

func (d *DBConnection) DeleteUserByID(userID string) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id=?", userID).Delete(&model.User{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id=?", userID).Delete(&model.PhoneOrder{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id=?", userID).Delete(&model.TextMessageOrder{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id=?", userID).Delete(&model.TextMessageDetail{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id=?", userID).Delete(&model.TextMessageReply{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id=?", userID).Delete(&model.UserRealNameAuthentication{}).Error; err != nil {
			return err
		}
		return nil
	})
}
