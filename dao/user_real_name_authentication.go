package dao

import (
	"phone_ios_backend/model"
	"time"
)

// CheckUserRealNameAuthenticationExist 根据UserName和IDCard查询是否存在记录
func (d *DBConnection) CheckUserRealNameAuthenticationExist(userID string) (bool, error) {
	var count int64
	if err := d.db.Model(&model.UserRealNameAuthentication{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// CreateUserRealNameAuthentication 创建实名认证记录
func (d *DBConnection) CreateUserRealNameAuthentication(userID string) error {
	var auth = &model.UserRealNameAuthentication{
		UserID:     userID,
		CreateTime: time.Now(),
	}
	// 创建记录
	if err := d.db.Create(auth).Error; err != nil {
		return err
	}
	return nil
}
