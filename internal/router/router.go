package router

import (
	"github.com/gin-gonic/gin"
	do "github.com/samber/do/v2"
)

func InitRouter(r *gin.Engine, i do.Injector) {
	initInjectorRouter(r, i)
	InitDemoRouter(r, i)
}
