package logger

import (
	"phone_ios_backend/common"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

func NewZapFileLogger(logPath string) *zap.SugaredLogger {
	logFileName := "wechat.log"
	path := filepath.Join(logPath, logFileName)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	writer := zapcore.AddSync(file)
	encoder := zapcore.NewJSONEncoder(config)
	core := zapcore.NewCore(encoder, writer, zap.NewAtomicLevelAt(zap.InfoLevel))
	// 创建 logger
	logger := zap.New(core)
	defer logger.Sync()
	sugar := logger.Sugar()
	return sugar
}

func NewZapLogger() {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	encoder := zapcore.NewJSONEncoder(config)
	core := zapcore.NewCore(encoder, os.Stdout, zap.NewAtomicLevelAt(zap.InfoLevel))
	// 创建 logger
	logger := zap.New(core)
	defer logger.Sync()
	sugar := logger.Sugar()
	Logger = sugar
}

type AppError struct {
	Code    int    // 错误码
	Message string // 错误信息
}

func (e *AppError) Error() string {
	return fmt.Sprintf("Code: %d, Message: %s", e.Code, e.Message)
}

// 构造函数：用于生成 AppError
func NewAppError(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func HandleError(c *gin.Context, err error) {
	// 如果是自定义错误类型 AppError
	var appErr *AppError
	if errors.As(err, &appErr) {
		c.JSON(http.StatusOK, common.RequestResp{
			Code:    appErr.Code,
			Message: appErr.Message,
		})
		return
	}
	// 如果是其他类型的错误，返回通用的错误
	c.JSON(http.StatusInternalServerError, common.RequestResp{
		Code:    5001,
		Message: err.Error(),
	})
}
