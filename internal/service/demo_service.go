package service

import (
	"context"
	"go_template/internal/pkg/zlog"

	v1 "go_template/api/v1"
	"go_template/internal/component"
	"go_template/internal/repository"

	"go.uber.org/zap"
)

type DemoService struct {
	demoRepo *repository.DemoRepository
	//xxxRepo *repository.XxxxRepository
	c *component.Component
}

func NewDemoService(
	demoRepo *repository.DemoRepository,
	c *component.Component,
) *DemoService {
	return &DemoService{
		demoRepo: demoRepo,
		c:        c,
	}
}

// Create 数据库事务的使用方法
func (ds *DemoService) Create(ctx context.Context, req *v1.AddAuthRequest) error {
	//rule := &model.Demo{
	//	// 赋值
	//}
	//
	return ds.c.Tm.Transaction(
		ctx, func(ctx context.Context) error {
			//if err := ds.demoRepo.Delete(ctx, 600006000); err != nil {
			//	ds.c.Log.C(ctx).Error("创建占位规则失败", zap.Error(err))
			//	if strings.Contains(err.Error(), constant.UniqueKey) {
			//		return errno.RuleExistsError
			//	}
			//	return errno.CreateRuleError
			//}

			//if err := ds.nodeRepo.BatchCreate(ctx, timeSlots); err != nil {
			//	s.c.Log.C(ctx).Error("创建占位规则时间段失败", zap.Error(err))
			//	if strings.Contains(err.Error(), constant.UniqueKey) {
			//		return errno.RuleExistsError
			//	}
			//	return errno.CreateRuleError
			//}
			return nil
		},
	)
}

// GetDemo 单表操作非数据库事务的使用方式
func (ds *DemoService) GetDemo(ctx context.Context, parkCode int64) (any, error) {
	zlog.C(ctx).Info("service", zap.Int64("park_code", parkCode))
	//rule, err := ds.demoRepo.GetByParkCode(ctx, parkCode)
	//if err != nil {
	//	return nil, err
	//}

	return nil, nil

}
