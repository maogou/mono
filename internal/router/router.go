package router

import (
	"go_template/internal/component"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine, c *component.Component) {
	InitDemoRouter(r, c)
}
