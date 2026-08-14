package service_test

import (
	"context"
	"testing"

	v1 "go_template/api/v1"
	"go_template/internal/pkg/zlog"
	"go_template/internal/service"
	mock_repository "go_template/test/mocks/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestDemoService_Create(t *testing.T) {
	must := require.New(t)

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockRepo := mock_repository.NewMockDemoRepository(ctrl)
	mockTx := mock_repository.NewMockTransaction(ctrl)

	mockTx.EXPECT().
		Transaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
			return fn(context.Background())
		}).
		Times(1)

	svc := service.NewDemoService(
		service.NewService(&zlog.Logger{Logger: zap.NewNop()}),
		mockRepo,
		mockTx,
	)

	must.NoError(svc.Create(context.Background(), &v1.AddAuthRequest{}))
}

func TestDemoService_GetDemo(t *testing.T) {
	is := assert.New(t)
	must := require.New(t)

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	svc := service.NewDemoService(
		service.NewService(&zlog.Logger{Logger: zap.NewNop()}),
		mock_repository.NewMockDemoRepository(ctrl),
		mock_repository.NewMockTransaction(ctrl),
	)

	// 在 ctx 放入 nop logger,避免 GetDemo 内部的 zlog.C(ctx) 落到默认 stdout logger
	ctx := context.WithValue(context.Background(), zlog.CtxLoggerKey, zap.NewNop())

	got, err := svc.GetDemo(ctx, 6000060000)
	must.NoError(err)
	is.Nil(got)
}
