package dao

import "phone_ios_backend/model"

func (d *DBConnection) CreatePhoneOrder(order *model.PhoneOrder) error {
	return d.db.Create(order).Error
}

func (d *DBConnection) GetPhoneOrderByID(id int) (*model.PhoneOrder, error) {
	var order model.PhoneOrder
	err := d.db.First(&order, id).Error
	return &order, err
}

func (d *DBConnection) UpdatePhoneOrder(order *model.PhoneOrder) error {
	return d.db.Save(order).Error
}

func (d *DBConnection) DeletePhoneOrder(id int) error {
	return d.db.Delete(&model.PhoneOrder{}, id).Error
}

func (d *DBConnection) GetAllPhoneOrdersByUserID(openid string) ([]model.PhoneOrder, error) {
	var orders []model.PhoneOrder
	err := d.db.Where("user_id=?", openid).Order("create_time desc").Find(&orders).Error
	return orders, err
}

//
// func (d *DBConnection) CreateUserOrder(order *model.UserOrder) error {
// 	return d.db.Create(order).Error
// }
//
// func (d *DBConnection) GetUserOrderByID(id int) (*model.UserOrder, error) {
// 	var order model.UserOrder
// 	err := d.db.First(&order, id).Error
// 	return &order, err
// }
//
// func (d *DBConnection) UpdateUserOrder(order *model.UserOrder) error {
// 	return d.db.Save(order).Error
// }
//
// func (d *DBConnection) DeleteUserOrder(id int) error {
// 	return d.db.Delete(&model.UserOrder{}, id).Error
// }
//
//
// func (d *DBConnection) GetAllUserOrders() ([]model.UserOrder, error) {
// 	var orders []model.UserOrder
// 	err := d.db.Find(&orders).Error
// 	return orders, err
// }
