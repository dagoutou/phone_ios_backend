package dao

import (
	"phone_ios_backend/connection"
	"phone_ios_backend/model"
	"gorm.io/gorm"
)

type IAPTransactionDAO struct {
	db *gorm.DB
}

func NewIAPTransactionDAO() *IAPTransactionDAO {
	return &IAPTransactionDAO{db: connection.DbConnection}
}

// ExistsByTransactionID 检查某苹果交易号是否已发放过(防重放)
func (d *IAPTransactionDAO) ExistsByTransactionID(txID string) (bool, error) {
	var cnt int64
	if err := d.db.Model(&model.IAPTransaction{}).
		Where("transaction_id = ?", txID).
		Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// CreateTx 在事务内写入 IAP 交易记录
func (d *IAPTransactionDAO) CreateTx(tx *gorm.DB, record *model.IAPTransaction) error {
	return tx.Create(record).Error
}
