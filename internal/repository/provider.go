package repository

import (
	"github.com/go-resty/resty/v2"
	do "github.com/samber/do/v2"
	"gorm.io/gorm"

	"go_template/internal/config"
	"go_template/internal/pkg/zlog"
)

func ProvideRepository(i do.Injector) (*Repository, error) {
	return NewRepository(
		do.MustInvoke[*gorm.DB](i),
		do.MustInvoke[*zlog.Logger](i),
	), nil
}

func ProvideThirdApi(i do.Injector) (*ThirdApi, error) {
	return NewThirdApi(
		do.MustInvoke[*zlog.Logger](i),
		do.MustInvoke[*config.Config](i),
		do.MustInvoke[*resty.Client](i),
	), nil
}

func ProvideTransaction(i do.Injector) (Transaction, error) {
	return NewTransaction(do.MustInvoke[*Repository](i)), nil
}

func ProvideDemoRepository(i do.Injector) (*DemoRepository, error) {
	return NewDemoRepository(do.MustInvoke[*Repository](i)), nil
}
