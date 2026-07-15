package command

import (
	"context"
	"errors"
	"fmt"

	"go_template/internal/component"
	"go_template/internal/config"
	"go_template/internal/pkg/httpc"
	"go_template/internal/pkg/zlog"
	"go_template/internal/repository"
	"go_template/internal/store"

	"github.com/urfave/cli/v3"
)

type AppCommand struct {
	comp component.Component
}

func NewApp() *cli.Command {
	cmd := &AppCommand{}

	return &cli.Command{
		Name:  "start",
		Usage: "start  http server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Usage: "config file path",
				Value: config.DefaultConfigPath,
			},
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			if err := cmd.initComponent(c.String("config")); err != nil {
				return ctx, err
			}
			return ctx, nil
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return run(&cmd.comp)
		},
		After: func(ctx context.Context, c *cli.Command) error {
			return cmd.comp.Shutdown()
		},
	}
}

func (cmd *AppCommand) initComponent(configPath string) error {
	cmd.comp.Conf = config.MustLoadConfig(configPath)
	cmd.comp.Log = zlog.NewZapLog(&cmd.comp.Conf)
	cmd.comp.Log.Info("Starting component initialization")

	if err := cmd.initDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	if err := cmd.initRedis(); err != nil {
		return fmt.Errorf("failed to initialize redis: %w", err)
	}

	if err := cmd.initHTTPClient(); err != nil {
		return fmt.Errorf("failed to initialize http client: %w", err)
	}

	if err := cmd.initRepository(); err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	cmd.comp.Log.Info("Component initialization completed")
	return nil
}

func (cmd *AppCommand) initDatabase() error {
	db, err := store.NewDB(&cmd.comp.Conf, cmd.comp.Log)
	if err != nil {
		return err
	}
	cmd.comp.Db = db
	cmd.comp.Log.Info("Database initialized successfully")
	return nil
}

func (cmd *AppCommand) initRedis() error {
	rdb, err := store.NewRedis(&cmd.comp.Conf)
	if err != nil {
		return err
	}

	cmd.comp.Rdb = rdb
	cmd.comp.Log.Info("Redis initialized successfully")

	return nil
}

func (cmd *AppCommand) initHTTPClient() error {
	cmd.comp.Client = httpc.NewClient(cmd.comp.Log)
	if cmd.comp.Client == nil {
		return errors.New("http client initialization returned nil")
	}
	cmd.comp.Log.Info("Resty HTTP client initialized successfully")
	return nil
}

func (cmd *AppCommand) initRepository() error {
	cmd.comp.Repository = repository.NewRepository(cmd.comp.Log, cmd.comp.Db)
	if cmd.comp.Repository == nil {
		return errors.New("repository initialization returned nil")
	}
	cmd.comp.Tm = repository.NewTransaction(cmd.comp.Repository)
	if cmd.comp.Tm == nil {
		return errors.New("transaction initialization returned nil")
	}

	cmd.comp.ThirdApi = repository.NewThirdApi(cmd.comp.Log, &cmd.comp.Conf, cmd.comp.Client)
	if cmd.comp.ThirdApi == nil {
		return errors.New("third api initialization returned nil")
	}
	cmd.comp.Log.Info("Repository initialized successfully")
	return nil
}
