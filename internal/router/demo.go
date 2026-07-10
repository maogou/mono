package router

import (
	"go_template/internal/component"
	"go_template/internal/repository"
	"go_template/internal/service"

	"github.com/gin-gonic/gin"

	"go_template/internal/handler"
)

func InitDemoRouter(r *gin.Engine, c *component.Component) {
	demoRepo := repository.NewDemoRepository(c.Repository)
	demoService := service.NewDemoService(demoRepo, c)
	d := handler.NewDemoHandler(demoService)
	r.GET("/", d.Health)
}
