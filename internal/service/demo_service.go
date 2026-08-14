package service

import (
	"context"

	v1 "go_template/api/v1"
	"go_template/internal/pkg/zlog"
	"go_template/internal/repository"

	"go.uber.org/zap"
)

type DemoService interface {
	Create(ctx context.Context, req *v1.AddAuthRequest) error
	GetDemo(ctx context.Context, parkCode int64) (any, error)
}

func NewDemoService(s *Service, demoRepo repository.DemoRepository, tm repository.Transaction) DemoService {
	return &demoService{
		demoRepo: demoRepo,
		tm:       tm,
		Service:  s,
	}
}

type demoService struct {
	demoRepo repository.DemoRepository
	tm       repository.Transaction
	*Service
}

func (ds *demoService) Create(ctx context.Context, req *v1.AddAuthRequest) error {
	return ds.tm.Transaction(
		ctx, func(ctx context.Context) error {
			return nil
		},
	)
}

func (ds *demoService) GetDemo(ctx context.Context, parkCode int64) (any, error) {
	zlog.C(ctx).Info("service", zap.Int64("park_code", parkCode))
	return nil, nil
}
