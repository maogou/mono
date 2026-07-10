package store

import (
	"context"
	"time"

	"go_template/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewRedis(conf *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(
		&redis.Options{
			Addr:     conf.Redis.Addr,
			Password: conf.Redis.Password,
			DB:       conf.Redis.DB,
		},
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.Redis.LockTimeout)*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	return rdb, nil
}
