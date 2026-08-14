package httpc

import (
	"github.com/go-resty/resty/v2"
	do "github.com/samber/do/v2"

	"go_template/internal/pkg/zlog"
)

func ProvideClient(i do.Injector) (*resty.Client, error) {
	return NewClient(do.MustInvoke[*zlog.Logger](i)), nil
}
