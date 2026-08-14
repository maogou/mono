package httpc

import (
	"time"

	"go_template/internal/pkg/zlog"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type ClientOptions struct {
	Timeout time.Duration
}

var DefaultOptions = ClientOptions{
	Timeout: 3 * time.Second,
}

func NewClient(logger *zlog.Logger) *resty.Client {
	return NewClientWithOptions(logger, DefaultOptions)
}

func NewClientWithOptions(logger *zlog.Logger, options ClientOptions) *resty.Client {
	client := resty.New()
	client.SetTimeout(options.Timeout)

	client.OnBeforeRequest(
		func(c *resty.Client, r *resty.Request) error {
			ctx := r.Context()

			logger.C(ctx).Info(
				"third-req-params",
				zap.String("method", r.Method),
				zap.String("url", c.BaseURL+r.URL),
				zap.Any("headers", r.Header),
				zap.Any("query", r.QueryParam),
				zap.Any("body", r.Body),
				zap.Time("start_time", time.Now()),
			)
			return nil
		},
	)

	client.OnAfterResponse(
		func(c *resty.Client, r *resty.Response) error {
			ctx := r.Request.Context()

			if r.Error() != nil {
				logger.C(ctx).Error(
					"third-res-error",
					zap.String("method", r.Request.Method),
					zap.String("url", r.Request.URL),
					zap.Int("status", r.StatusCode()),
					zap.Any("error", r.Error()),
					zap.String("cost", r.Time().String()),
				)
			} else {
				logger.C(ctx).Info(
					"third-res-body",
					zap.String("method", r.Request.Method),
					zap.String("url", r.Request.URL),
					zap.Int("status", r.StatusCode()),
					zap.ByteString("body", r.Body()),
					zap.String("cost", r.Time().String()),
				)
			}
			return nil
		},
	)

	return client
}
