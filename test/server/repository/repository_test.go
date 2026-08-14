package repository_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"go_template/internal/model"
	"go_template/internal/pkg/zlog"
	"go_template/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// newTestRepository 构建基于内存 SQLite 的测试仓库。
func newTestRepository(t *testing.T) *repository.Repository {
	t.Helper()
	must := require.New(t)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	must.NoError(err, "open sqlite")

	sqlDB, err := db.DB()
	must.NoError(err, "get sql.DB")
	// 内存 SQLite 每个连接一个独立库,限制单连接保证 AutoMigrate 的表对所有查询可见
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	must.NoError(db.AutoMigrate(&model.Demo{}), "migrate")

	return repository.NewRepository(db, &zlog.Logger{Logger: zap.NewNop()})
}

// newDemo 插入一条测试数据,email 按 parkCode 区分以满足唯一索引。
func newDemo(t *testing.T, r *repository.Repository, parkCode int64) {
	t.Helper()
	err := r.Tx(context.Background()).Create(
		&model.Demo{
			Name:     "alice",
			Email:    fmt.Sprintf("user%d@example.com", parkCode),
			Password: "secret",
			ParkCode: parkCode,
		},
	).Error
	require.NoError(t, err, "seed")
}

func TestRepository_Transaction(t *testing.T) {
	t.Run("commit", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		r := newTestRepository(t)
		ctx := context.Background()

		err := r.Transaction(ctx, func(ctx context.Context) error {
			return r.Tx(ctx).Create(
				&model.Demo{
					Name:     "bob",
					Email:    "bob@example.com",
					Password: "secret",
					ParkCode: 1,
				},
			).Error
		})
		must.NoError(err)

		var count int64
		must.NoError(r.Tx(ctx).Model(&model.Demo{}).Count(&count).Error)
		is.Equal(int64(1), count)
	})

	t.Run("rollback", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		r := newTestRepository(t)
		ctx := context.Background()

		err := r.Transaction(ctx, func(ctx context.Context) error {
			_ = r.Tx(ctx).Create(
				&model.Demo{
					Name:     "bob",
					Email:    "bob@example.com",
					Password: "secret",
					ParkCode: 1,
				},
			).Error
			return errors.New("boom")
		})
		must.Error(err)

		var count int64
		must.NoError(r.Tx(ctx).Model(&model.Demo{}).Count(&count).Error)
		is.Zero(count)
	})
}

// 事务内 r.Tx(ctx) 应返回事务连接:能读到未提交的数据。
func TestRepository_Transaction_TxContext(t *testing.T) {
	is := assert.New(t)
	must := require.New(t)

	r := newTestRepository(t)
	ctx := context.Background()

	err := r.Transaction(ctx, func(ctx context.Context) error {
		if err := r.Tx(ctx).Create(
			&model.Demo{
				Name:     "carol",
				Email:    "carol@example.com",
				Password: "secret",
				ParkCode: 2,
			},
		).Error; err != nil {
			return err
		}
		var count int64
		if err := r.Tx(ctx).Model(&model.Demo{}).Count(&count).Error; err != nil {
			return err
		}
		is.Equal(int64(1), count, "事务内应能读到未提交数据")
		return nil
	})
	must.NoError(err)
}
