package store

import (
	"context"
	"time"

	"go_template/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(conf *config.Redis) *redis.Client {
	return redis.NewClient(
		&redis.Options{
			Addr:     conf.Addr,
			Password: conf.Password,
			DB:       conf.DB,
		},
	)
}

func PingRedis(rdb *redis.Client, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return rdb.Ping(ctx).Err()
}
