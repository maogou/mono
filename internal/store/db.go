package store

import (
	"time"

	"go_template/internal/config"
	"go_template/internal/pkg/zapgorm"
	"go_template/internal/pkg/zlog"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDB(conf *config.Config, l *zlog.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(
		mysql.Open(conf.DB.Dsn), &gorm.Config{
			Logger: zapgorm.New(l.Logger),
		},
	)
	if err != nil {
		return nil, err
	}
	db = db.Debug()

	sqlDB, err := db.DB()
	if err != nil {
		return db, err
	}
	sqlDB.SetMaxIdleConns(conf.DB.MaxIdleConns)
	sqlDB.SetMaxOpenConns(conf.DB.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(conf.DB.ConnMaxLifetime) * time.Second)
	return db, nil
}
