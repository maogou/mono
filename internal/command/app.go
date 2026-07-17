package command

import (
	"context"

	"go_template/internal/config"
	"go_template/internal/pkg/zlog"

	"github.com/redis/go-redis/v9"
	do "github.com/samber/do/v2"
	"github.com/urfave/cli/v3"
	"gorm.io/gorm"
)

type AppCommand struct {
	di do.Injector
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
			if err := cmd.initInjector(c.String("config")); err != nil {
				return ctx, err
			}
			return ctx, nil
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return run(cmd.di)
		},
		After: func(ctx context.Context, c *cli.Command) error {
			return cmd.shutdown()
		},
	}
}

func (cmd *AppCommand) shutdown() error {
	db, _ := do.Invoke[*gorm.DB](cmd.di)
	if db != nil {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}

	rdb, _ := do.Invoke[*redis.Client](cmd.di)
	if rdb != nil {
		_ = rdb.Close()
	}

	logger, _ := do.Invoke[*zlog.Logger](cmd.di)
	if logger != nil {
		_ = logger.Sync()
	}
	return nil
}
