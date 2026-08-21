package api

import (
	"phone_ios_backend/common"
	"phone_ios_backend/config"
	"phone_ios_backend/dao"
	"phone_ios_backend/logger"
	"net/http"
	
	"github.com/gin-gonic/gin"
)

func GetPhonePlan(ctx *gin.Context) {
	PhoneCodes := []string{"70000"}
	if config.Setting.APP.IsOnLine {
		PhoneCodes = []string{"80000", "90000", "20000"}
	}
	// params := model.NewPhoneNumberRequestParamsV2()
	// if err := ctx.ShouldBindJSON(&params); err != nil {
	// 	logger.HandleError(ctx, err)
	// 	return
	// }
	// resp, err := my_utils.SendPostRequestUrl("https://xx.shizaihao.top/api/findTaoCan", params)
	// if err != nil {
	// 	logger.HandleError(ctx, err)
	// 	return
	// }
	// var res Response
	// if err := json.Unmarshal(resp, &res); err != nil {
	// 	logger.HandleError(ctx, err)
	// 	return
	// }
	// taoCan := res.Data.ApiTaoCanVo
	// var tempCode []string
	// for _, s := range taoCan {
	// 	tempCode = append(tempCode, s.PhoneCode)
	// }
	// really := Intersect(PhoneCodes, tempCode)
	db := dao.NewDB()
	phonePlans, err := db.GetPhonePlansByCodes(PhoneCodes)
	if err != nil {
		logger.HandleError(ctx, err)
		return
	}
	if !config.Setting.APP.IsOnLine {
		for i := 0; i < len(phonePlans); i++ {
			phonePlans[i].PhoneName = "一个隐私号，1小时有效。"
		}
	}
	mp := make(map[string]interface{})
	mp["phone_plan"] = phonePlans
	ctx.JSON(http.StatusOK, common.RequestResp{
		Code:    http.StatusOK,
		Data:    mp,
		Message: "success",
	})
}

type Response struct {
	Data struct {
		ResCode     string `json:"resCode"`
		ResMsg      string `json:"resMsg"`
		Sign        any    `json:"sign"`
		ApiTaoCanVo []struct {
			PhoneCode string `json:"phoneCode"`
			PhoneName string `json:"phoneName"`
		} `json:"aApiTaoCanVo"`
	} `json:"data"`
	ResMsg string `json:"resMsg"`
	Status string `json:"status"`
}

func Intersect(a, b []string) []string {
	set := make(map[string]struct{}, len(a))
	for _, v := range a {
		set[v] = struct{}{}
	}
	
	var res []string
	for _, v := range b {
		if _, ok := set[v]; ok {
			res = append(res, v)
		}
	}
	return res
}

func GetPriceMenus(ctx *gin.Context) {
	db := dao.NewDB()
	menus, err := db.GetPriceMenu()
	if err != nil {
		logger.HandleError(ctx, err)
		return
	}
	mp := make(map[string]interface{})
	mp["menus"] = menus
	ctx.JSON(http.StatusOK, common.RequestResp{
		Code:    http.StatusOK,
		Data:    mp,
		Message: "success",
	})
}

func GetIsOnLine(ctx *gin.Context) {
	mp := make(map[string]interface{})
	mp["is_online"] = config.Setting.APP.IsOnLine
	ctx.JSON(http.StatusOK, common.RequestResp{
		Code:    http.StatusOK,
		Data:    mp,
		Message: "success",
	})
}

// GetProductPrice 根据商品名称查询价格
func GetProductPrice(ctx *gin.Context) {
	productName := ctx.Query("name")
	if productName == "" {
		err := logger.NewAppError(http.StatusBadRequest, "name is required")
		logger.HandleError(ctx, err)
		return
	}

	dao := dao.NewProductPriceDAO()
	productPrice, err := dao.GetProductPriceByName(productName)
	if err != nil {
		logger.HandleError(ctx, err)
		return
	}

	mp := make(map[string]interface{})
	mp["product_price"] = productPrice
	ctx.JSON(http.StatusOK, common.RequestResp{
		Code:    http.StatusOK,
		Data:    mp,
		Message: "success",
	})
}

// GetAllProductPrices 获取所有商品价格列表
func GetAllProductPrices(ctx *gin.Context) {
	dao := dao.NewProductPriceDAO()
	productPrices, err := dao.GetAllProductPrices()
	if err != nil {
		logger.HandleError(ctx, err)
		return
	}

	mp := make(map[string]interface{})
	mp["product_prices"] = productPrices
	ctx.JSON(http.StatusOK, common.RequestResp{
		Code:    http.StatusOK,
		Data:    mp,
		Message: "success",
	})
}
