package command

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"go_template/internal/config"
	"go_template/internal/middleware"
	"go_template/internal/pkg/zlog"
	"go_template/internal/router"

	"github.com/gin-gonic/gin"
	do "github.com/samber/do/v2"
	"go.uber.org/zap"
)

func run(i do.Injector) error {
	conf := do.MustInvoke[*config.Config](i)
	logger := do.MustInvoke[*zlog.Logger](i)

	gin.SetMode(conf.Mode)
	route := gin.New()

	addr := ":" + strconv.Itoa(conf.Port)
	route.Use(
		gin.CustomRecovery(middleware.CustomRecovery(logger)),
		middleware.RequestLog(logger), middleware.ResponseLog(logger),
	)
	router.InitRouter(route, i)

	srv := &http.Server{
		Addr:              addr,
		Handler:           route,
		ReadHeaderTimeout: time.Duration(conf.ReadTimeout) * time.Second,
	}

	logger.Info("http-api服务访问地址==>http://127.0.0.1" + addr)
	logger.Info("终止服务,请按键盘上 Ctrl+C 键退出服务")

	srvErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srvErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logger.Info("收到系统信号,准备关闭服务", zap.String("signal", sig.String()))
	case err := <-srvErr:
		logger.Error("监听"+addr+"端口失败", zap.Error(err))
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.ShutdownTimeout)*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Warn("http-api服务异常,关闭失败")
		return err
	}

	logger.Info("已终止http-api对外接口访问")

	return nil
}
