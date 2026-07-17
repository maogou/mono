package store

import (
	"time"

	"go_template/internal/config"
	"go_template/internal/pkg/zapgorm"
	"go_template/internal/pkg/zlog"

	do "github.com/samber/do/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDB(i do.Injector) (*gorm.DB, error) {
	conf := do.MustInvoke[*config.Config](i)
	logger := do.MustInvoke[*zlog.Logger](i)

	db, err := gorm.Open(
		mysql.Open(conf.DB.Dsn), &gorm.Config{
			Logger: zapgorm.New(logger.Logger),
		},
	)
	if err != nil {
		return nil, err
	}
	db = db.Debug()

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(conf.DB.MaxIdleConns)
	sqlDB.SetMaxOpenConns(conf.DB.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(conf.DB.ConnMaxLifetime) * time.Second)
	return db, nil
}
