package repository

import (
	"context"

	do "github.com/samber/do/v2"

	"go_template/internal/model"
)

type DemoRepository struct {
	*Repository
}

func NewDemoRepository(i do.Injector) (*DemoRepository, error) {
	r := do.MustInvoke[*Repository](i)
	return &DemoRepository{
		Repository: r,
	}, nil
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
