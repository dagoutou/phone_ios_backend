package dao

import (
	"phone_ios_backend/connection"
	"phone_ios_backend/model"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserPointsDAO struct {
	db *gorm.DB
}

func NewUserPointsDAO() *UserPointsDAO {
	return &UserPointsDAO{db: connection.DbConnection}
}

// GetByUser 查询用户积分记录;若不存在返回零值对象(不报错),由调用方决定是否需要初始化
func (d *UserPointsDAO) GetByUser(userID, appName string) (*model.UserPoints, error) {
	var p model.UserPoints
	err := d.db.Where("user_id = ? AND app_name = ?", userID, appName).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &model.UserPoints{UserID: userID, AppName: appName, Balance: 0}, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetByUserTx 在事务内读取并加行锁;不存在返回 nil(由调用方初始化)
func (d *UserPointsDAO) GetByUserTx(tx *gorm.DB, userID, appName string) (*model.UserPoints, error) {
	var p model.UserPoints
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND app_name = ?", userID, appName).
		First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Increase 在事务内增加积分并写明细。amount 必须为正数,返回发放后的新余额。
// 调用方需自己开启 tx,本函数不再起事务。
func (d *UserPointsDAO) Increase(tx *gorm.DB, userID, appName string, amount int, source, refOrderNo, productID, remark string) (int, error) {
	if amount <= 0 {
		return 0, errors.New("amount must be positive")
	}
	existing, err := d.GetByUserTx(tx, userID, appName)
	if err != nil {
		return 0, err
	}
	var newBalance int
	if existing == nil {
		newBalance = amount
		row := &model.UserPoints{
			UserID: userID, AppName: appName, Balance: newBalance,
		}
		if err = tx.Create(row).Error; err != nil {
			return 0, err
		}
	} else {
		newBalance = existing.Balance + amount
		if err = tx.Model(&model.UserPoints{}).
			Where("user_id = ? AND app_name = ?", userID, appName).
			Update("balance", newBalance).Error; err != nil {
			return 0, err
		}
	}
	detail := &model.PointTransaction{
		UserID:       userID,
		AppName:      appName,
		Type:         model.PointTypeRecharge,
		Amount:       amount,
		BalanceAfter: newBalance,
		Source:       source,
		RefOrderNo:   refOrderNo,
		ProductID:    productID,
		Remark:       remark,
	}
	if err = tx.Create(detail).Error; err != nil {
		return 0, err
	}
	return newBalance, nil
}

// ListTransactionsByUser 分页查询用户积分明细
func (d *UserPointsDAO) ListTransactionsByUser(userID, appName string, page, size int) ([]model.PointTransaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	var total int64
	if err := d.db.Model(&model.PointTransaction{}).
		Where("user_id = ? AND app_name = ?", userID, appName).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.PointTransaction
	if err := d.db.Where("user_id = ? AND app_name = ?", userID, appName).
		Order("id DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Decrease 在事务内扣减积分并写明细。amount 必须为正数,返回扣减后的新余额。
// 调用方需自己开启 tx,本函数不再起事务。
func (d *UserPointsDAO) Decrease(tx *gorm.DB, userID, appName string, amount int, source, refOrderNo, productID, remark string) (int, error) {
	if amount <= 0 {
		return 0, errors.New("amount must be positive")
	}
	existing, err := d.GetByUserTx(tx, userID, appName)
	if err != nil {
		return 0, err
	}
	if existing == nil {
		return 0, errors.New("user points not found")
	}
	if existing.Balance < amount {
		return 0, errors.New("insufficient balance")
	}
	newBalance := existing.Balance - amount
	if err = tx.Model(&model.UserPoints{}).
		Where("user_id = ? AND app_name = ?", userID, appName).
		Update("balance", newBalance).Error; err != nil {
		return 0, err
	}
	detail := &model.PointTransaction{
		UserID:       userID,
		AppName:      appName,
		Type:         model.PointTypeConsume,
		Amount:       -amount,
		BalanceAfter: newBalance,
		Source:       source,
		RefOrderNo:   refOrderNo,
		ProductID:    productID,
		Remark:       remark,
	}
	if err = tx.Create(detail).Error; err != nil {
		return 0, err
	}
	return newBalance, nil
}
