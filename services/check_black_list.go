package services

import (
	"phone_ios_backend/dao"
)

func CheckBlacklist(phoneNumber string) (bool, error) {
	db := dao.NewDB()
	inBlacklist, err := db.IsPhoneNumberInBlacklist(phoneNumber)
	if err != nil {
		return false, err
	}
	return inBlacklist, nil
}
