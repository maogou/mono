package command

import (
	"fmt"
	"github.com/samber/do/v2"

	"go_template/internal/config"
	"go_template/internal/pkg/httpc"
	"go_template/internal/pkg/zlog"
	"go_template/internal/repository"
	"go_template/internal/service"
	"go_template/internal/store"
)

func (cmd *AppCommand) initInjector(configPath string) error {
	conf, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	injector := do.New()

	do.ProvideValue(injector, &conf)

	do.Provide(injector, zlog.NewZapLog)
	do.Provide(injector, store.NewDB)
	do.Provide(injector, store.NewRedis)
	do.Provide(injector, httpc.NewClient)
	do.Provide(injector, repository.NewRepository)
	do.Provide(injector, repository.NewThirdApi)
	do.Provide(injector, repository.NewTransaction)
	do.Provide(injector, repository.NewDemoRepository)
	do.Provide(injector, service.NewDemoService)

	cmd.di = injector

	logger := do.MustInvoke[*zlog.Logger](injector)
	logger.Info("Component initialization completed")
	return nil
}
