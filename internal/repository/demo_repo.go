package repository

import (
	"context"

	"go_template/internal/model"
)

type DemoRepository struct {
	*Repository
}

func NewDemoRepository(r *Repository) *DemoRepository {
	return &DemoRepository{
		Repository: r,
	}
}

// Delete 如果你定义的这方方法是在数据库事务中使用, 请你使用 Tx(ctx) 方法
func (d *DemoRepository) Delete(ctx context.Context, parkCode int64) error {
	return d.Tx(ctx).Where("park_code = ? ", parkCode).Delete(&model.Demo{}).Error
}

// GetByParkCode 如果你定义的这个方法不会被在事务中使用那么建议你使用d.db.WithContext(ctx) 方法
func (d *DemoRepository) GetByParkCode(ctx context.Context, parkCode int64) ([]model.Demo, error) {
	var demos []model.Demo
	if err := d.db.WithContext(ctx).Where("park_code = ?", parkCode).Find(&demos).Error; err != nil {
		return nil, err
	}
	return demos, nil
}
