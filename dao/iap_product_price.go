package dao

import (
	"phone_ios_backend/connection"
	"phone_ios_backend/model"
	"errors"

	"gorm.io/gorm"
)

type IAPProductPriceDAO struct {
	db *gorm.DB
}

func NewIAPProductPriceDAO() *IAPProductPriceDAO {
	return &IAPProductPriceDAO{db: connection.DbConnection}
}

// GetByProductID 根据 App Store product_id 查询商品配置
func (d *IAPProductPriceDAO) GetByProductID(productID string) (*model.IAPProductPrice, error) {
	var p model.IAPProductPrice
	if err := d.db.Where("product_id = ?", productID).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// ListAll 查询所有 IAP 商��,按价格升序
func (d *IAPProductPriceDAO) ListAll() ([]model.IAPProductPrice, error) {
	var list []model.IAPProductPrice
	if err := d.db.Order("price ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// GetByName 根据内部名称查询商品配置
func (d *IAPProductPriceDAO) GetByName(name string) (*model.IAPProductPrice, error) {
	var p model.IAPProductPrice
	if err := d.db.Where("name = ?", name).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// GetByProductIDTx 在事务内查询商品配置
func (d *IAPProductPriceDAO) GetByProductIDTx(tx *gorm.DB, productID string) (*model.IAPProductPrice, error) {
	var p model.IAPProductPrice
	if err := tx.Where("product_id = ?", productID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}
