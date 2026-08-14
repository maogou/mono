package zlog

import (
	do "github.com/samber/do/v2"

	"go_template/internal/config"
)

func ProvideZapLog(i do.Injector) (*Logger, error) {
	return NewZapLog(do.MustInvoke[*config.Config](i)), nil
}
