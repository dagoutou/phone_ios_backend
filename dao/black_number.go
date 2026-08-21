package dao

func (d *DBConnection) IsPhoneNumberInBlacklist(phoneNumber string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM black_numbers WHERE phone_number = ?"
	err := d.db.Raw(query, phoneNumber).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
