package help

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maogou/period"
	"github.com/rs/xid"

	"go_template/internal/constant"
)

func TraceId(ctx context.Context) string {
	if c, ok := ctx.(*gin.Context); ok {
		if qid := c.GetString(constant.Qid); qid != "" {
			return qid
		}
	}

	if v := ctx.Value(constant.Qid); v != nil {
		if qid, ok := v.(string); ok && qid != "" {
			return qid
		}
	}

	return xid.New().String()
}

func CeilDiv[T int | int64](a, b T) T {
	if a <= 0 || b <= 0 {
		return 0
	}

	return (a + b - 1) / b
}

func FormatPeriods(p []period.Period) string {
	var result string
	for _, v := range p {
		result += v.Format(time.DateTime) + ","
	}
	return strings.TrimRight(result, ",")
}

func BuildTime(date time.Time, t time.Time) time.Time {
	return time.Date(
		date.Year(), date.Month(), date.Day(),
		t.Hour(), t.Minute(), t.Second(), 0,
		time.Local,
	)
}

func Adjust235959(t time.Time) time.Time {
	if t.Format(time.TimeOnly) == constant.Time235959 {
		return t.Add(1 * time.Second)
	}
	return t
}

func StartZeroDate(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}

func FormatAmount(amount int64) string {
	if amount <= 0 {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", float64(amount)/100)
}
