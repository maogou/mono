package service

import (
	do "github.com/samber/do/v2"

	"go_template/internal/pkg/zlog"
	"go_template/internal/repository"
)

func ProvideService(i do.Injector) (*Service, error) {
	return NewService(do.MustInvoke[*zlog.Logger](i)), nil
}

func ProvideDemoService(i do.Injector) (DemoService, error) {
	return NewDemoService(
		do.MustInvoke[*Service](i),
		do.MustInvoke[repository.DemoRepository](i),
		do.MustInvoke[repository.Transaction](i),
	), nil
}
