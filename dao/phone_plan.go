package dao

import "phone_ios_backend/model"

func (d *DBConnection) GetPhonePlansByCodes(phoneCodes []string) ([]model.PhonePlan, error) {
	var plans []model.PhonePlan
	// 使用Where条件查询phoneCode在给定数组中的所有记录
	result := d.db.Where("phone_code IN ?", phoneCodes).Find(&plans)
	if result.Error != nil {
		return nil, result.Error
	}

	return plans, nil
}
