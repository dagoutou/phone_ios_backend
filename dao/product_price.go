package dao

import (
	"phone_ios_backend/connection"
	"phone_ios_backend/model"
	"gorm.io/gorm"
)

type ProductPriceDAO struct {
	db *gorm.DB
}

func NewProductPriceDAO() *ProductPriceDAO {
	return &ProductPriceDAO{
		db: connection.DbConnection,
	}
}

// GetProductPriceByName 根据商品名称查询价格
func (d *ProductPriceDAO) GetProductPriceByName(name string) (*model.ProductPrice, error) {
	var productPrice model.ProductPrice
	result := d.db.Where("name = ?", name).First(&productPrice)
	if result.Error != nil {
		return nil, result.Error
	}
	return &productPrice, nil
}

// GetAllProductPrices 获取所有商品价格
func (d *ProductPriceDAO) GetAllProductPrices() ([]model.ProductPrice, error) {
	var productPrices []model.ProductPrice
	result := d.db.Find(&productPrices)
	if result.Error != nil {
		return nil, result.Error
	}
	return productPrices, nil
}
