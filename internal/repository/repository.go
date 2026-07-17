package repository

import (
	"context"

	"github.com/go-resty/resty/v2"
	do "github.com/samber/do/v2"

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

func NewThirdApi(i do.Injector) (*ThirdApi, error) {
	logger := do.MustInvoke[*zlog.Logger](i)
	conf := do.MustInvoke[*config.Config](i)
	client := do.MustInvoke[*resty.Client](i)
	return &ThirdApi{
		logger: logger,
		client: client,
		conf:   conf,
	}, nil
}

func NewRepository(i do.Injector) (*Repository, error) {
	logger := do.MustInvoke[*zlog.Logger](i)
	db := do.MustInvoke[*gorm.DB](i)
	return &Repository{
		db:     db,
		logger: logger,
	}, nil
}

type Transaction interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

func NewTransaction(i do.Injector) (Transaction, error) {
	r := do.MustInvoke[*Repository](i)
	return r, nil
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
