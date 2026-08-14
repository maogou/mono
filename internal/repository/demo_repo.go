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

func (d *DemoRepository) Delete(ctx context.Context, parkCode int64) error {
	return d.Tx(ctx).Where("park_code = ? ", parkCode).Delete(&model.Demo{}).Error
}

func (d *DemoRepository) GetByParkCode(ctx context.Context, parkCode int64) ([]model.Demo, error) {
	var demos []model.Demo
	if err := d.db.WithContext(ctx).Where("park_code = ?", parkCode).Find(&demos).Error; err != nil {
		return nil, err
	}
	return demos, nil
}
