package router

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
	dohttp "github.com/samber/do/v2/http"
	"net/http"
)

func initInjectorRouter(r *gin.Engine, i do.Injector) {
	basePath := "/di"
	di := r.Group(basePath)
	di.GET(
		"/", func(c *gin.Context) {
			html, err := dohttp.IndexHTML(basePath)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, html)
		},
	)

	di.GET(
		"/scope", func(c *gin.Context) {
			scopeID := c.Query("scope_id")
			html, err := dohttp.ScopeTreeHTML(basePath, i, scopeID)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, html)
		},
	)

	di.GET(
		"/service", func(c *gin.Context) {
			scopeID := c.Query("scope_id")
			serviceName := c.Query("service_name")

			if serviceName != "" {
				html, err := dohttp.ServiceHTML(basePath, i, scopeID, serviceName)
				if err != nil {
					c.String(http.StatusInternalServerError, err.Error())
					return
				}
				c.Header("Content-Type", "text/html; charset=utf-8")
				c.String(http.StatusOK, html)
			} else {
				html, err := dohttp.ServiceListHTML(basePath, i)
				if err != nil {
					c.String(http.StatusInternalServerError, err.Error())
					return
				}
				c.Header("Content-Type", "text/html; charset=utf-8")
				c.String(http.StatusOK, html)
			}
		},
	)
}
