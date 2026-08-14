package repository_test

import (
	"context"
	"testing"

	"go_template/internal/model"
	"go_template/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDemoRepository_GetByParkCode(t *testing.T) {
	tests := []struct {
		name     string
		seed     []int64
		parkCode int64
		wantLen  int
	}{
		{
			name:     "found",
			seed:     []int64{42},
			parkCode: 42,
			wantLen:  1,
		},
		{
			name:     "empty",
			seed:     []int64{42},
			parkCode: 1,
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			must := require.New(t)

			r := newTestRepository(t)
			for _, pc := range tt.seed {
				newDemo(t, r, pc)
			}
			repo := repository.NewDemoRepository(r)

			got, err := repo.GetByParkCode(context.Background(), tt.parkCode)
			must.NoError(err)
			is.Len(got, tt.wantLen)
		})
	}
}

func TestDemoRepository_Delete(t *testing.T) {
	is := assert.New(t)
	must := require.New(t)

	r := newTestRepository(t)
	newDemo(t, r, 42)
	newDemo(t, r, 43)
	repo := repository.NewDemoRepository(r)

	must.NoError(repo.Delete(context.Background(), 42))

	ctx := context.Background()
	var count int64
	must.NoError(r.Tx(ctx).Model(&model.Demo{}).Where("park_code = ?", 42).Count(&count).Error)
	is.Zero(count, "park_code=42 的数据应全部删除")

	must.NoError(r.Tx(ctx).Model(&model.Demo{}).Where("park_code = ?", 43).Count(&count).Error)
	is.Equal(int64(1), count, "park_code=43 的数据不应受影响")
}
