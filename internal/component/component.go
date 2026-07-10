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
