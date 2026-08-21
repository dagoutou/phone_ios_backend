package api

import (
	"phone_ios_backend/common"
	"phone_ios_backend/dao"
	"phone_ios_backend/logger"
	"phone_ios_backend/my_utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// func WechatPay(ctx *gin.Context) {
// 	var req struct {
// 		Openid  string `json:"openid"`
// 		AppName string `json:"app_name"`
// 	}
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
// 		return
// 	}
// 	params := config.CreateOrder(req.Openid, req.AppName)
// 	resp, _, err := config.Pay(params, req.AppName)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"PrepayWithRequestPayment error:": err})
// 		return
// 	}
// 	mp := make(map[string]string)
// 	mp["timeStamp"] = *resp.TimeStamp
// 	mp["nonceStr"] = *resp.NonceStr
// 	mp["package"] = *resp.Package
// 	mp["signType"] = *resp.SignType
// 	mp["paySign"] = *resp.PaySign
// 	mp["appid"] = *resp.Appid
// 	mp["outTradeNo"] = *params.OutTradeNo
// 	ctx.JSON(http.StatusOK, gin.H{"data": mp})
// }
//
// func CreateUserOrder(ctx *gin.Context) {
// 	var userOrder model.UserOrder
// 	if err := ctx.ShouldBindJSON(&userOrder); err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"bind query error:": err})
// 		return
// 	}
// 	db := dao.NewDB()
// 	if err := db.CreateUserOrder(&userOrder); err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"CreatePhoneOrder error:": err})
// 		return
// 	}
// 	ctx.JSON(http.StatusOK, gin.H{"data": userOrder.ID})
// }
//
// func NotifyChengLu(ctx *gin.Context) {
// 	notify, err := config.CreateNotify(config.ChengLU)
// 	if err != nil {
// 		logger.Logger.Errorf("CreateNotify failed error:%v", err)
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"CreateNotify failed error:": err})
// 		return
// 	}
// 	transaction := new(payments.Transaction)
// 	notifyReq, err := notify.ParseNotifyRequest(context.Background(), ctx.Request, transaction)
// 	if err != nil {
// 		logger.Logger.Errorf("Signature verification failed error:%v", err)
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"Signature verification failed error:": err})
// 		return
// 	}
// 	ctx.JSON(http.StatusOK, gin.H{"notifyReq": notifyReq})
// }
//
// func NotifyMiaoMiao(ctx *gin.Context) {
// 	notify, err := config.CreateNotify(config.MiaoMiao)
// 	if err != nil {
// 		logger.Logger.Errorf("CreateNotify failed error:%v", err)
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"CreateNotify failed error:": err})
// 		return
// 	}
// 	transaction := new(payments.Transaction)
// 	notifyReq, err := notify.ParseNotifyRequest(context.Background(), ctx.Request, transaction)
// 	if err != nil {
// 		logger.Logger.Errorf("Signature verification failed error:%v", err)
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"Signature verification failed error:": err})
// 		return
// 	}
// 	ctx.JSON(http.StatusOK, gin.H{"notifyReq": notifyReq})
// }
//
// func NotifyBaby(ctx *gin.Context) {
// 	notify, err := config.CreateNotify(config.Baby)
// 	if err != nil {
// 		logger.Logger.Errorf("CreateNotify failed error:%v", err)
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"CreateNotify failed error:": err})
// 		return
// 	}
// 	transaction := new(payments.Transaction)
// 	notifyReq, err := notify.ParseNotifyRequest(context.Background(), ctx.Request, transaction)
// 	if err != nil {
// 		logger.Logger.Errorf("Signature verification failed error:%v", err)
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"Signature verification failed error:": err})
// 		return
// 	}
// 	ctx.JSON(http.StatusOK, gin.H{"notifyReq": notifyReq})
// }
//
// func NotifyKuaiHu(ctx *gin.Context) {
// 	notify, err := config.CreateNotify(config.KuaiHu)
// 	if err != nil {
// 		logger.Logger.Errorf("CreateNotify failed error:%v", err)
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"CreateNotify failed error:": err})
// 		return
// 	}
// 	transaction := new(payments.Transaction)
// 	notifyReq, err := notify.ParseNotifyRequest(context.Background(), ctx.Request, transaction)
// 	if err != nil {
// 		logger.Logger.Errorf("Signature verification failed error:%v", err)
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"Signature verification failed error:": err})
// 		return
// 	}
// 	ctx.JSON(http.StatusOK, gin.H{"notifyReq": notifyReq})
// }
//
// func NotifyAiAi(ctx *gin.Context) {
// 	notify, err := config.CreateNotify(config.Aiai)
// 	if err != nil {
// 		logger.Logger.Errorf("CreateNotify failed error:%v", err)
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"CreateNotify failed error:": err})
// 		return
// 	}
// 	transaction := new(payments.Transaction)
// 	notifyReq, err := notify.ParseNotifyRequest(context.Background(), ctx.Request, transaction)
// 	if err != nil {
// 		logger.Logger.Errorf("Signature verification failed error:%v", err)
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"Signature verification failed error:": err})
// 		return
// 	}
// 	ctx.JSON(http.StatusOK, gin.H{"notifyReq": notifyReq})
// }
//
// func CreatePhoneOrder(ctx *gin.Context) {
// 	var phoneOrder model.PhoneOrder
// 	if err := ctx.ShouldBindJSON(&phoneOrder); err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"bind query error:": err})
// 		return
// 	}
// 	db := dao.NewDB()
// 	if err := db.CreatePhoneOrder(&phoneOrder); err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"CreatePhoneOrder error:": err})
// 		return
// 	}
// 	ctx.JSON(http.StatusOK, gin.H{"success": "success"})
// }

func RealNameAuthentication(ctx *gin.Context) {
	var req struct {
		UserName  string `json:"user_name"`
		IDCard    string `json:"id_card"`
		UserID    string `json:"user_id"`
		OpenID    string `json:"open_id"`
		FaceImage string `json:"face_image"`
	}
	var resp struct {
		Valid   bool   `json:"valid"`
		Count   int64  `json:"count"`
		Message string `json:"message"`
		Code    int    `json:"code"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = logger.NewAppError(http.StatusBadRequest, err.Error())
		logger.HandleError(ctx, err)
		return
	}
	count, err := my_utils.GetRealNameAuthCount(req.UserID)
	if err != nil {
		logger.HandleError(ctx, err)
		return
	}
	resp.Count = count
	if count == 0 {
		err = logger.NewAppError(4004, "今日实名认证次数已达上限，请联系客服！")
		logger.HandleError(ctx, err)
		return
	}
	authentication, err := my_utils.AliyunRealName(req.UserName, req.IDCard, req.FaceImage)
	if err != nil {
		logger.HandleError(ctx, err)
		return
	}
	resp.Message = authentication.Desc
	resp.Code = authentication.Code
	if authentication.Code != 0 {
		err = logger.NewAppError(4005, authentication.Desc)
		logger.HandleError(ctx, err)
		return
	}
	db := dao.NewDB()
	if err = db.CreateUserRealNameAuthentication(req.UserID); err != nil {
		logger.HandleError(ctx, err)
		return
	}
	resp.Valid = true
	ctx.JSON(http.StatusOK, common.RequestResp{
		Code:    http.StatusOK,
		Data:    resp,
		Message: "success",
	})
}

func GetRealName(ctx *gin.Context) {
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = logger.NewAppError(http.StatusBadRequest, err.Error())
		logger.HandleError(ctx, err)
		return
	}
	db := dao.NewDB()
	exist, err := db.CheckUserRealNameAuthenticationExist(req.UserID)
	if err != nil {
		logger.HandleError(ctx, err)
		return
	}
	mp := make(map[string]interface{})
	mp["valid"] = false
	if !exist {
		err = logger.NewAppError(4006, "您未进行实名认证！")
		logger.HandleError(ctx, err)
		return
	}
	mp["valid"] = true
	ctx.JSON(http.StatusOK, common.RequestResp{
		Code:    http.StatusOK,
		Data:    mp,
		Message: "success",
	})
}

// func CreateUserAuth(ctx *gin.Context) {
// 	var req struct {
// 		UserName    string `json:"user_name"`
// 		IDCard      string `json:"id_card"`
// 		PhoneNumber string `json:"phone_number"`
// 	}
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
// 		return
// 	}
// 	db := dao.NewDB()
// 	if err := db.CreateUserRealNameAuthentication(req.PhoneNumber, req.UserName, req.UserName); err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"CreateUserRealNameAuthentication error:": err})
// 		return
// 	}
// 	ctx.JSON(http.StatusOK, gin.H{"data": "success"})
// }
