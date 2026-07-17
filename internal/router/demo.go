package router

import (
	"github.com/gin-gonic/gin"
	do "github.com/samber/do/v2"

	"go_template/internal/handler"
	"go_template/internal/service"
)

func InitDemoRouter(r *gin.Engine, i do.Injector) {
	demoService := do.MustInvoke[*service.DemoService](i)
	d := handler.NewDemoHandler(demoService)
	r.GET("/", d.Health)
}
