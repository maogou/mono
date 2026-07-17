package middleware

import (
	"bytes"
	"io"
	"slices"
	"time"

	"go_template/internal/constant"
	"go_template/internal/pkg/zlog"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

type RequestParams struct {
	Method  string              `json:"method"`
	Url     string              `json:"url"`
	Headers map[string][]string `json:"headers"`
	Params  string              `json:"request_params"`
}

var noWriteBodyPath = []string{
	"/di",
	"/di/scope",
	"/di/service",
}

func RequestLog(logger *zlog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		qid := xid.New().String()
		ctx.Set(constant.Qid, qid)
		logger.V(ctx, zap.String(constant.Qid, qid))

		var requestParams = RequestParams{
			Method:  ctx.Request.Method,
			Url:     ctx.Request.URL.String(),
			Headers: ctx.Request.Header,
		}
		if ctx.Request.Body != nil {
			if bodyBytes, err := ctx.GetRawData(); err == nil {
				ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				requestParams.Params = string(bodyBytes)
			} else {
				requestParams.Params = "read body failed"
				logger.C(ctx).Error("获取请求体失败", zap.Error(err))
			}

		}
		logger.C(ctx).Info("Request", zap.Any("request_params", requestParams))
		ctx.Writer.Header().Set(constant.Qid, qid)
		ctx.Next()
	}
}
func ResponseLog(logger *zlog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
		ctx.Writer = blw
		startTime := time.Now()
		ctx.Next()
		duration := time.Since(startTime).String()
		if slices.Contains(noWriteBodyPath, ctx.Request.URL.Path) {
			logger.C(ctx).Info(
				"Response", zap.Any("time", duration),
			)
		} else {
			logger.C(ctx).Info(
				"Response", zap.String("response_body", blw.body.String()), zap.Any("time", duration),
			)
		}

	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	if _, err := w.body.Write(b); err != nil {
		return 0, err
	}
	return w.ResponseWriter.Write(b)
}
