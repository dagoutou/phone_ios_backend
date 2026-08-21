package router

import (
	"fmt"
	"net/http"
	"os"
	"phone_ios_backend/api"
	"phone_ios_backend/common"
	"phone_ios_backend/my_utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func Router(g *gin.Engine) {
	// 允许跨域请求（CORS）
	g.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // 允许所有来源访问
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Account-No, X-Phone, X-Sign")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})
	g.Static("/images", "./static/images")
	g.StaticFile("/sms_reply_upload.html", "./static/sms_reply_upload.html")
	g.StaticFile("/sms_order_process.html", "./static/sms_order_process.html")
	html := g.Group("/static")
	miaomiao := html.Group("/miaomiao")
	{
		miaomiao.GET("/using_tutorials", func(c *gin.Context) {
			c.HTML(http.StatusOK, "using_tutorials.html", gin.H{
				"title": "使用教程",
			})
		})
		miaomiao.GET("/user_agreement", func(c *gin.Context) {
			c.HTML(http.StatusOK, "user_agreement.html", gin.H{
				"title": "用户协议",
			})
		})
		miaomiao.GET("/privacy_agreement", func(c *gin.Context) {
			c.HTML(http.StatusOK, "privacy_agreement.html", gin.H{
				"title": "隐私协议",
			})
		})
		miaomiao.GET("/faq", func(c *gin.Context) {
			c.HTML(http.StatusOK, "faq.html", gin.H{
				"title": "faq",
			})
		})
		miaomiao.GET("/describe", func(c *gin.Context) {
			c.HTML(http.StatusOK, "describe.html", gin.H{
				"title": "describe",
			})
		})
	}
	xiaolu := html.Group("/xiaolu")
	{
		xiaolu.GET("/using_tutorials", func(c *gin.Context) {
			data, err := os.ReadFile("./static/xiaolu/using_tutorials.txt")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("无法读取文件: %v", err)})
				return
			}
			mp := make(map[string]interface{})
			mp["text"] = string(data)
			c.JSON(http.StatusOK, common.RequestResp{
				Code:    http.StatusOK,
				Data:    mp,
				Message: "success",
			})
		})
		xiaolu.GET("/user_agreement", func(c *gin.Context) {
			data, err := os.ReadFile("./static/xiaolu/user_agreement.txt")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("无法读取文件: %v", err)})
				return
			}
			c.JSON(200, gin.H{
				"view": "container",
				"text": string(data),
			})
		})
		xiaolu.GET("/privacy_agreement", func(c *gin.Context) {
			data, err := os.ReadFile("./static/xiaolu/privacy_agreement.txt")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("无法读取文件: %v", err)})
				return
			}
			c.JSON(200, gin.H{
				"view": "container",
				"text": string(data),
			})
		})
		xiaolu.GET("/faq", func(c *gin.Context) {
			data, err := os.ReadFile("./static/xiaolu/faq.txt")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("无法读取文件: %v", err)})
				return
			}
			c.JSON(200, gin.H{
				"view": "container",
				"text": string(data),
			})
		})
		xiaolu.GET("/describe", func(c *gin.Context) {
			data, err := os.ReadFile("./static/xiaolu/describe.txt")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("无法读取文件: %v", err)})
				return
			}
			c.JSON(200, gin.H{
				"view": "container",
				"text": string(data),
			})
		})
	}
	g.Use(AuthMiddleware())
	group := g.Group("/v2")
	{
		group.POST("/login", api.LoginHandler)
		group.POST("/logout", api.Logout)
		group.POST("/send_code", api.SendSMSCode)
		group.GET("/protected", api.ProtectedAPI)
		group.POST("/get_phone_plans", api.GetPhonePlan)
		group.GET("/get_price_menus", api.GetPriceMenus)
		group.GET("/is_online", api.GetIsOnLine)
		group.GET("/product_price", api.GetProductPrice)
		group.GET("/product_prices", api.GetAllProductPrices)
		// group.GET("/black_phone", api.CheckBlacklist)
		user_auth := group.Group("")
		{
			user_auth.POST("/real_name", api.RealNameAuthentication)
			user_auth.POST("/get_user", api.GetRealName)
		}
		pay := group.Group("")
		{
			pay.POST("/payments/apple_iap/verify", api.AppleIAPVerify)
			pay.POST("/phone_check", api.PhoneCheck)
		}
		points := group.Group("/points")
		{
			points.GET("/balance", api.GetPointsBalance)
			points.GET("/transactions", api.GetPointTransactions)
			points.GET("/iap/products", api.GetAppleIAPProducts)
			points.POST("/payment", api.PointsPayment)
		}
		// 外部API路由，使用独立的认证中间件
		external := group.Group("/external")
		{
			// external.Use(api.ExternalAuthMiddleware())
			external.GET("/plans", api.GetExternalPlans)
			external.POST("/orders", api.CreateExternalOrder)
			external.GET("/orders/detail", api.GetExternalOrderDetail)
			external.GET("/orders/list", api.GetExternalOrderList)
			external.GET("/balance", api.GetExternalBalance)
		}
		// 短信相关路由
		sms := group.Group("/sms")
		{
			sms.POST("/send", api.SendTextMessage)
			sms.POST("/reply/upload", api.UploadTextMessageReplyImage)
			sms.POST("/order/complete", api.CompleteTextMessageOrder)
			sms.POST("/order/reject", api.RejectTextMessageOrder)
			sms.GET("/orders", api.GetTextMessageOrders)
			sms.GET("/details", api.GetTextMessageDetails)
			sms.GET("/order/detail", api.GetTextMessageOrderDetail)
			sms.GET("/replies", api.GetSMSReplies)
			sms.GET("/replies/by_order", api.GetSMSRepliesByOrderNo)
		}
		// 短信回调（不需要认证）
		g.POST("/pull_sms_upstream", api.PullSMSUpstream)
		g.POST("/pull_sms_status", api.PullSMSStatus)
	}
	// g.POST("/v4/payments/xiaolu/wechat/callback", api.WechatPayXiaoLuCallback) // TODO: 处理函数已删除,待重新实现
}

// AuthMiddleware JWT 验证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 判断是否为登录接口，不是则进行 Token 验证
		if c.Request.URL.Path == "/v2/login" || c.Request.URL.Path == "/v2/request" || c.Request.URL.Path == "/v2/black_phone" || c.Request.URL.Path == "/v2/aliPay/login" || strings.Contains(c.Request.URL.Path, "images") ||
			strings.Contains(c.Request.URL.Path, "/v2") || strings.Contains(c.Request.URL.Path, "/v2/send_code") || strings.Contains(c.Request.URL.Path, "/v2/payments") || strings.Contains(c.Request.URL.Path, "/v4/payments") ||
			c.Request.URL.Path == "/pull_sms_upstream" || c.Request.URL.Path == "/pull_sms_status" {

			// 如果是登录接口，跳过验证
			c.Next() // 继续处理请求
			return
		}
		// 获取请求头中的 Authorization 字段
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供Token"})
			c.Abort() // 阻止后续处理
			return
		}

		// Authorization 格式：Bearer <token>
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 格式不正确"})
			c.Abort()
			return
		}

		// 解析 Token
		claims, err := my_utils.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// 将解析后的 claims 存储到 context 中，以便后续处理
		c.Set("claims", claims)
		c.Next()
	}
}
