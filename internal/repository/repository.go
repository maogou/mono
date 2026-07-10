package repository

import (
	"context"

	"github.com/go-resty/resty/v2"

	"go_template/internal/config"
	"go_template/internal/pkg/zlog"

	"gorm.io/gorm"
)

type ctxKey string

const ctxTxKey ctxKey = "TxKey"

type Repository struct {
	db     *gorm.DB
	logger *zlog.Logger
}

type ThirdApi struct {
	logger *zlog.Logger
	client *resty.Client
	conf   *config.Config
}

func NewThirdApi(logger *zlog.Logger, conf *config.Config, client *resty.Client) *ThirdApi {
	return &ThirdApi{
		logger: logger,
		client: client,
		conf:   conf,
	}
}

func NewRepository(logger *zlog.Logger, db *gorm.DB) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

type Transaction interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

func NewTransaction(r *Repository) Transaction {
	return r
}

func (r *Repository) Tx(ctx context.Context) *gorm.DB {
	v := ctx.Value(ctxTxKey)
	if v != nil {
		if tx, ok := v.(*gorm.DB); ok {
			return tx
		}
	}
	return r.db.WithContext(ctx)
}

func (r *Repository) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			ctx = context.WithValue(ctx, ctxTxKey, tx)
			return fn(ctx)
		},
	)
}
