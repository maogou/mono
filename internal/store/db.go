package store

import (
	"time"

	"go_template/internal/config"
	"go_template/internal/pkg/zapgorm"
	"go_template/internal/pkg/zlog"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDB(conf *config.DB, logger *zlog.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(
		mysql.Open(conf.Dsn), &gorm.Config{
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
	sqlDB.SetMaxIdleConns(conf.MaxIdleConns)
	sqlDB.SetMaxOpenConns(conf.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Second)
	return db, nil
}
