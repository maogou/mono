package repository

import (
	"context"

	"go_template/internal/model"
)

type DemoRepository interface {
	Delete(ctx context.Context, parkCode int64) error
	GetByParkCode(ctx context.Context, parkCode int64) ([]model.Demo, error)
}

func NewDemoRepository(r *Repository) DemoRepository {
	return &demoRepository{
		Repository: r,
	}
}

type demoRepository struct {
	*Repository
}

func (d *demoRepository) Delete(ctx context.Context, parkCode int64) error {
	return d.Tx(ctx).Where("park_code = ?", parkCode).Delete(&model.Demo{}).Error
}

func (d *demoRepository) GetByParkCode(ctx context.Context, parkCode int64) ([]model.Demo, error) {
	var demos []model.Demo
	if err := d.Tx(ctx).Where("park_code = ?", parkCode).Find(&demos).Error; err != nil {
		return nil, err
	}
	return demos, nil
}
