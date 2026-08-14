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
	do.Provide(cmd.di, zlog.ProvideZapLog)
	do.Provide(cmd.di, store.ProvideDB)
	do.Provide(cmd.di, store.ProvideRedis)
	do.Provide(cmd.di, httpc.ProvideClient)
	do.Provide(cmd.di, repository.ProvideRepository)
	do.Provide(cmd.di, repository.ProvideThirdApi)
	do.Provide(cmd.di, repository.ProvideTransaction)
	do.Provide(cmd.di, repository.ProvideDemoRepository)
	do.Provide(cmd.di, service.ProvideService)
	do.Provide(cmd.di, service.ProvideDemoService)

	logger := do.MustInvoke[*zlog.Logger](cmd.di)
	logger.Info("Component initialization completed")
}
