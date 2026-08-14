package service

import (
	"context"

	v1 "go_template/api/v1"
	"go_template/internal/pkg/zlog"
	"go_template/internal/repository"

	"go.uber.org/zap"
)

type DemoService struct {
	demoRepo *repository.DemoRepository
	tm       repository.Transaction
	log      *zlog.Logger
}

func NewDemoService(demoRepo *repository.DemoRepository, tm repository.Transaction, log *zlog.Logger) *DemoService {
	return &DemoService{
		demoRepo: demoRepo,
		tm:       tm,
		log:      log,
	}
}

func (ds *DemoService) Create(ctx context.Context, req *v1.AddAuthRequest) error {
	return ds.tm.Transaction(
		ctx, func(ctx context.Context) error {
			return nil
		},
	)
}

func (ds *DemoService) GetDemo(ctx context.Context, parkCode int64) (any, error) {
	zlog.C(ctx).Info("service", zap.Int64("park_code", parkCode))
	return nil, nil
}
