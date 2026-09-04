package command

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"go_template/internal/config"
	"go_template/internal/constant"
	"go_template/internal/middleware"
	"go_template/internal/pkg/updater"
	"go_template/internal/pkg/zlog"
	"go_template/internal/router"

	"github.com/gin-gonic/gin"
	do "github.com/samber/do/v2"
	"go.uber.org/zap"
)

func newEngine(conf *config.Config, logger *zlog.Logger, i do.Injector) *gin.Engine {
	gin.SetMode(conf.Mode)
	route := gin.New()
	route.Use(
		gin.CustomRecovery(middleware.CustomRecovery(logger)),
		middleware.RequestLog(logger), middleware.ResponseLog(logger),
	)
	router.InitRouter(route, i)
	return route
}

func run(i do.Injector) error {
	conf := do.MustInvoke[*config.Config](i)
	logger := do.MustInvoke[*zlog.Logger](i)

	route := newEngine(conf, logger, i)

	addr := ":" + strconv.Itoa(conf.Port)
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

	restartCh := make(chan struct{}, 1)
	uCtx, uCancel := context.WithCancel(context.Background())
	defer uCancel()

	restarting := false
	var restarter *updater.Restarter
	if conf.Update != nil && conf.Update.Enable {
		restarter = updater.NewRestarter()
		u := updater.New(
			updater.Options{
				CheckIntervalSeconds:   conf.Update.CheckInterval,
				FirstCheckDelaySeconds: 15,               // 启动后错开一段再首轮检查
				CurrentVersion:         constant.Version, // 版本号内建,发版=改常量后重编译
				CheckURL:               conf.Update.CheckURL,
				DownloadTimeoutSeconds: conf.Update.DownloadTimeout,
				MaxSizeMB:              conf.Update.MaxSizeMB,
				Restarter:              restarter, // 注入同实例,apply 成功后即 Commit 正式路径
			}, logger.Logger,
		)
		go u.Run(uCtx, restartCh)
		logger.Info("自升级已开启,更新源==>" + conf.Update.CheckURL + "  当前版本号为 " + constant.Version)
	}

	select {
	case sig := <-quit:
		logger.Info("收到系统信号,准备关闭服务", zap.String("signal", sig.String()))
	case <-restartCh:
		restarting = true
		logger.Info("自升级已提交,准备关闭服务后重启(Linux/macOS 同 PID exec;Windows 拉起新实例)")
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

	if restarting && restarter != nil {
		logger.Info("开始执行自动重启升级")
		_ = shutdownResources(i)
		if err := restarter.Exec(); err != nil {
			logger.Error("自升级重启失败,交由外部守护拉起新二进制", zap.Error(err))
			return fmt.Errorf("自升级重启失败(exec): %w", err)
		}
	}
	return nil
}
