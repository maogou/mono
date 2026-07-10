package handler

import (
	"go_template/internal/pkg/zlog"
	"go_template/internal/service"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go_template/internal/pkg/response"
)

type DemoHandler struct {
	ds *service.DemoService
}

func NewDemoHandler(ds *service.DemoService) *DemoHandler {
	return &DemoHandler{
		ds: ds,
	}
}

func (d *DemoHandler) Health(ctx *gin.Context) {
	zlog.C(ctx).Info("aaa", zap.String("aa", "cc"))
	_, _ = d.ds.GetDemo(ctx, 6000060000)
	response.Success(
		ctx,
		gin.H{"tip": "年轻的时候, 谁没有点乱七八糟的爱情那?", "time": time.Now().Format(time.DateTime), "email": "kinyou_xy@foxmail.com"},
	)
}
