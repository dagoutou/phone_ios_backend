package api

import (
	"io"
	"net/http"

	"phone_ios_backend/config"
	"phone_ios_backend/logger"
	"phone_ios_backend/my_utils"

	"github.com/gin-gonic/gin"
)

// ExternalAuthMiddleware 校验 X-Account-No、X-Phone、X-Sign 请求头
func ExternalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		accountNo := c.GetHeader("X-Account-No")
		phone := c.GetHeader("X-Phone")
		sign := c.GetHeader("X-Sign")

		if accountNo == "" || phone == "" || sign == "" {
			c.JSON(http.StatusUnauthorized, my_utils.ExternalResp{
				Code: 401,
				Msg:  "missing authentication headers",
			})
			c.Abort()
			return
		}

		ext := config.Setting.APP.External
		if accountNo != ext.AccountNo || phone != ext.Phone || sign != ext.Sign {
			c.JSON(http.StatusUnauthorized, my_utils.ExternalResp{
				Code: 401,
				Msg:  "invalid authentication",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func proxyGet(c *gin.Context, path string) {
	result, statusCode, err := my_utils.SendExternalGet(path, c.Request.URL.Query())
	if err != nil {
		logger.Logger.Errorf("external proxyGet failed: %v", err)
		c.JSON(http.StatusInternalServerError, my_utils.ExternalResp{
			Code: 500,
			Msg:  "upstream request failed: " + err.Error(),
		})
		return
	}
	c.JSON(statusCode, result)
}

func proxyPost(c *gin.Context, path string) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, my_utils.ExternalResp{
			Code: 400,
			Msg:  "read request body failed",
		})
		return
	}
	result, statusCode, err := my_utils.SendExternalPost(path, body)
	if err != nil {
		logger.Logger.Errorf("external proxyPost failed: %v", err)
		c.JSON(http.StatusInternalServerError, my_utils.ExternalResp{
			Code: 500,
			Msg:  "upstream request failed: " + err.Error(),
		})
		return
	}
	c.JSON(statusCode, result)
}

// GetExternalPlans GET /external/plans
// @Summary 获取外部套餐列表
// @Description 代理转发至上游 /external/plans，返回可用套餐列表
// @Tags external
// @Produce json
// @Success 200 {object} my_utils.ExternalResp "上游返回的套餐列表"
// @Failure 500 {object} my_utils.ExternalResp "上游请求或解析失败"
// @Router /external/plans [get]
func GetExternalPlans(c *gin.Context) {
	proxyGet(c, "/external/plans")
}

// CreateExternalOrder POST /external/orders
// @Summary 创建外部订单
// @Description 代理转发至上游 /external/orders，请求体原样透传，由上游定义字段契约
// @Tags external
// @Accept json
// @Produce json
// @Param order body object true "订单创建参数（字段由上游定义，原样透传）"
// @Success 200 {object} my_utils.ExternalResp "上游返回的订单创建结果"
// @Failure 400 {object} my_utils.ExternalResp "读取请求体失败"
// @Failure 500 {object} my_utils.ExternalResp "上游请求或解析失败"
// @Router /external/orders [post]
func CreateExternalOrder(c *gin.Context) {
	proxyPost(c, "/external/orders")
}

// GetExternalOrderDetail GET /external/orders/detail
// @Summary 获取外部订单详情
// @Description 代理转发至上游 /external/orders/detail，查询参数原样透传
// @Tags external
// @Produce json
// @Param orderNo query string true "订单号"
// @Success 200 {object} my_utils.ExternalResp "上游返回的订单详情"
// @Failure 500 {object} my_utils.ExternalResp "上游请求或解析失败"
// @Router /external/orders/detail [get]
func GetExternalOrderDetail(c *gin.Context) {
	proxyGet(c, "/external/orders/detail")
}

// GetExternalOrderList GET /external/orders/list
// @Summary 获取外部订单列表
// @Description 代理转发至上游 /external/orders/list，查询参数原样透传
// @Tags external
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} my_utils.ExternalResp "上游返回的订单列表"
// @Failure 500 {object} my_utils.ExternalResp "上游请求或解析失败"
// @Router /external/orders/list [get]
func GetExternalOrderList(c *gin.Context) {
	proxyGet(c, "/external/orders/list")
}

// GetExternalBalance GET /external/balance
// @Summary 获取外部账户余额
// @Description 代理转发至上游 /external/balance，返回当前账户余额
// @Tags external
// @Produce json
// @Success 200 {object} my_utils.ExternalResp "上游返回的余额信息"
// @Failure 500 {object} my_utils.ExternalResp "上游请求或解析失败"
// @Router /external/balance [get]
func GetExternalBalance(c *gin.Context) {
	proxyGet(c, "/external/balance")
}
