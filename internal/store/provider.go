package store

import (
	"time"

	do "github.com/samber/do/v2"

	"go_template/internal/config"
	"go_template/internal/pkg/zlog"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func ProvideDB(i do.Injector) (*gorm.DB, error) {
	cfg := do.MustInvoke[*config.Config](i)
	return NewDB(cfg.DB, do.MustInvoke[*zlog.Logger](i))
}

func ProvideRedis(i do.Injector) (*redis.Client, error) {
	cfg := do.MustInvoke[*config.Config](i)
	rdb := NewRedisClient(cfg.Redis)
	if err := PingRedis(rdb, time.Duration(cfg.Redis.LockTimeout)*time.Second); err != nil {
		_ = rdb.Close()
		return nil, err
	}
	return rdb, nil
}
