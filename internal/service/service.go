package service

import "go_template/internal/pkg/zlog"

type Service struct {
	logger *zlog.Logger
}

func NewService(logger *zlog.Logger) *Service {
	return &Service{
		logger: logger,
	}
}
