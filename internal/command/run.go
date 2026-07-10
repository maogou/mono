package command

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"go_template/internal/component"
	"go_template/internal/middleware"
	"go_template/internal/router"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func run(c *component.Component) error {
	gin.SetMode(c.Conf.Mode)
	route := gin.New()

	addr := ":" + strconv.Itoa(c.Conf.Port)
	route.Use(
		gin.CustomRecovery(middleware.CustomRecovery(c.Log)),
		middleware.RequestLog(c.Log), middleware.ResponseLog(c.Log),
	)
	router.InitRouter(route, c)

	srv := &http.Server{
		Addr:              addr,
		Handler:           route,
		ReadHeaderTimeout: 5 * time.Second,
	}

	c.Log.Info("http-api服务访问地址==>http://127.0.0.1" + addr)
	c.Log.Info("终止服务,请按键盘上 Ctrl+C 键退出服务")

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			c.Log.Fatal("监听"+addr+"端口失败", zap.Error(err))
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		c.Log.Warn("http-api服务异常,关闭失败")
		return err
	}

	c.Log.Info("已终止http-api对外接口访问")

	return nil
}
