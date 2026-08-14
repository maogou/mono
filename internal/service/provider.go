package service

import (
	do "github.com/samber/do/v2"

	"go_template/internal/pkg/zlog"
	"go_template/internal/repository"
)

func ProvideDemoService(i do.Injector) (*DemoService, error) {
	return NewDemoService(
		do.MustInvoke[*repository.DemoRepository](i),
		do.MustInvoke[repository.Transaction](i),
		do.MustInvoke[*zlog.Logger](i),
	), nil
}
