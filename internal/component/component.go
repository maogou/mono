package component

import (
	"go_template/internal/config"
	"go_template/internal/pkg/zlog"
	"go_template/internal/repository"

	"github.com/go-resty/resty/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Component struct {
	Conf       config.Config
	Log        *zlog.Logger
	Db         *gorm.DB
	Rdb        *redis.Client
	Client     *resty.Client
	Repository *repository.Repository
	Tm         repository.Transaction
	ThirdApi   *repository.ThirdApi
}

func (c *Component) Shutdown() error {
	c.Log.Info("关闭数据库连接和刷新日志")
	if c.Db != nil {
		if sqlDB, err := c.Db.DB(); err == nil {
			c.Log.Info("关闭数据库连接")
			_ = sqlDB.Close()
		}
	}
	if c.Rdb != nil {
		c.Log.Info("关闭redis连接")
		_ = c.Rdb.Close()
	}
	if c.Log != nil {
		c.Log.Info("刷新日志")
		_ = c.Log.Sync()
	}
	return nil
}
