package middleware

import (
	"github.com/gin-gonic/gin"

	"go_template/internal/constant"
	"go_template/internal/pkg/zlog"

	"go.uber.org/zap"
)

func CustomRecovery(logger *zlog.Logger) func(*gin.Context, any) {
	return func(ctx *gin.Context, err any) {
		logger.C(ctx).Error("Server panic", zap.Any("error", err))

		ctx.JSON(500, gin.H{"code": 500, "message": "服务器内部错误,请稍后再试!", "qid": ctx.GetString(constant.Qid)})
	}
}
