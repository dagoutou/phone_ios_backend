package api

import (
	"phone_ios_backend/common"
	"phone_ios_backend/config/sms"
	"phone_ios_backend/dao"
	"phone_ios_backend/logger"
	"phone_ios_backend/my_utils"
	"phone_ios_backend/services"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func LoginHandler(c *gin.Context) {
	var req struct {
		AppName string `json:"app_name"`
		Code    string `json:"code"`
		UserID  string `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		err = logger.NewAppError(http.StatusBadRequest, err.Error())
		logger.HandleError(c, err)
		return
	}
	if req.Code == "" || req.UserID == "" {
		err := logger.NewAppError(http.StatusBadRequest, "验证码/手机号不能为空")
		logger.HandleError(c, err)
		return
	}
	blacklist, err := services.CheckBlacklist(req.UserID)
	if err != nil {
		logger.HandleError(c, err)
		return
	}
	if blacklist {
		err := logger.NewAppError(http.StatusUnauthorized, "您暂时无法登录APP请联系客服！")
		logger.HandleError(c, err)
		return

	}
	code, err := my_utils.GetPhoneCode(req.UserID)
	if err != nil || code != req.Code {
		err := logger.NewAppError(4001, "验证码错误！")
		logger.HandleError(c, err)
		return
	}
	// 生成 JWT Token
	token, err := my_utils.GenerateJWT(req.UserID)
	if err != nil {
		logger.HandleError(c, err)
		return
	}
	// 判断用户是否存在，不存在存入数据库
	d := dao.NewDB()
	exists, m, err := d.GetUserByID(req.UserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.HandleError(c, err)
		return
	}
	if !exists {
		user, err := d.CreateUser(req.UserID)
		if err != nil {
			logger.HandleError(c, err)
			return
		}
		m = user
	}
	mp := make(map[string]interface{})
	mp["token"] = token
	mp["user"] = m
	c.JSON(http.StatusOK, common.RequestResp{
		Code:    http.StatusOK,
		Data:    mp,
		Message: "success",
	})
}

func Logout(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		err = logger.NewAppError(http.StatusBadRequest, err.Error())
		logger.HandleError(c, err)
		return
	}
	d := dao.NewDB()
	if err := d.DeleteUserByID(req.UserID); err != nil {
		logger.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, common.RequestResp{
		Code:    http.StatusOK,
		Message: "success",
	})
}

func SendSMSCode(ctx *gin.Context) {
	var req struct {
		PhoneNumber string `json:"phone_number"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = logger.NewAppError(http.StatusBadRequest, err.Error())
		logger.HandleError(ctx, err)
		return
	}
	if req.PhoneNumber == "" {
		err := logger.NewAppError(http.StatusBadRequest, "手机号不能为空")
		logger.HandleError(ctx, err)
		return
	}
	mp := make(map[string]interface{})
	mp["message"] = "验证码发送失败！"
	mp["code"] = 0
	canSend, err := my_utils.CanSendCode(req.PhoneNumber)
	if err != nil {
		logger.HandleError(ctx, err)
		return
	}
	if !canSend {
		// mp["message"] = "今日验证码发送次数已达上限！"
		// mp["code"] = 3
		err := logger.NewAppError(4002, "今日验证码发送次数已达上限！")
		logger.HandleError(ctx, err)
		return
	}
	code, err := my_utils.GetPhoneCode(req.PhoneNumber)
	if err != nil {
		logger.HandleError(ctx, err)
		return
	}
	if code != "" {
		// mp["message"] = "验证码还在有效期内！"
		// mp["code"] = 2
		err = logger.NewAppError(4003, "验证码还在有效期内！")
		logger.HandleError(ctx, err)
		return
	}
	if err = sms.SendSMSCode(req.PhoneNumber); err != nil {
		logger.HandleError(ctx, err)
		return
	}
	mp["message"] = "验证码发送成功,有效期5分钟！"
	mp["code"] = 1
	ctx.JSON(http.StatusOK, common.RequestResp{
		Code: http.StatusOK,
		Data: mp,
	})
}

func ProtectedAPI(c *gin.Context) {
	openID, _ := c.Get("openid")
	c.JSON(http.StatusOK, gin.H{"message": "Token 校验成功", "openID": openID})
}
