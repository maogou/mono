package command

import (
	"github.com/samber/do/v2"

	"go_template/internal/config"
	"go_template/internal/pkg/httpc"
	"go_template/internal/pkg/zlog"
	"go_template/internal/repository"
	"go_template/internal/service"
	"go_template/internal/store"
)

func (cmd *AppCommand) initInjector(conf config.Config) {
	do.ProvideValue(cmd.di, &conf)
	do.Provide(cmd.di, zlog.NewZapLog)
	do.Provide(cmd.di, store.NewDB)
	do.Provide(cmd.di, store.NewRedis)
	do.Provide(cmd.di, httpc.NewClient)
	do.Provide(cmd.di, repository.NewRepository)
	do.Provide(cmd.di, repository.NewThirdApi)
	do.Provide(cmd.di, repository.NewTransaction)
	do.Provide(cmd.di, repository.NewDemoRepository)
	do.Provide(cmd.di, service.NewDemoService)

	logger := do.MustInvoke[*zlog.Logger](cmd.di)
	logger.Info("Component initialization completed")
}
