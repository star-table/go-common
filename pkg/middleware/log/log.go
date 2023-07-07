package log

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/metadata"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	pkgErrors "github.com/star-table/go-common/pkg/errors"
	"go.opentelemetry.io/otel/trace"
)

var logNoPrint log.Level = -2

// LogServerMiddleware 记录服务端日志，目前只记录错误日志
func LogServerMiddleware(logger log.Logger, isDebug bool) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			var (
				code      int32
				reason    string
				kind      string
				operation string
			)
			startTime := time.Now()
			if info, ok := transport.FromServerContext(ctx); ok {
				kind = info.Kind().String()
				operation = info.Operation()
			}
			reply, err = handler(ctx, req)
			causeErr := pkgErrors.Cause(err)
			if se := errors.FromError(causeErr); se != nil {
				code = se.Code
				reason = se.Reason
			}
			metaInfo, _ := metadata.FromServerContext(ctx)
			level, stack := extractError(err, isDebug)
			if (level != logNoPrint || isDebug) && !strings.Contains(operation, "ping") {
				sc := trace.SpanContextFromContext(ctx)
				keys := make([]interface{}, 0, 10)
				keys = append(keys,
					"trace_id", sc.TraceID().String(),
					"kind", "server",
					"operation", operation,
					"component", kind,
					"args", extractArgs(req),
					"meta", metaInfo,
					"code", code,
					"reason", reason,
					"stack", stack,
					"latency", time.Since(startTime).Seconds(),
				)
				if errs := pkgErrors.Msg(err); len(errs) > 0 {
					keys = append(keys, "error_msg", errs)
				}
				_ = log.WithContext(ctx, logger).Log(level, keys...)
			}

			err = causeErr

			return
		}
	}
}

// extractError returns the string of the error
func extractError(err error, isDebug bool) (log.Level, string) {
	type levelInterface interface {
		GetLevel() log.Level
	}
	type stackTracer interface {
		StackTrace() pkgErrors.StackTrace
	}
	if err != nil {
		stack := ""
		if s, ok := err.(stackTracer); ok {
			stack = fmt.Sprintf("%+v", s.StackTrace())
		} else {
			stack = fmt.Sprintf("%+v", err)
		}
		if l, ok := err.(levelInterface); ok {
			return l.GetLevel(), stack
		}
		return log.LevelError, stack
	}
	if isDebug {
		return log.LevelInfo, ""
	}
	return logNoPrint, ""
}

// extractArgs returns the string of the req
func extractArgs(req interface{}) string {
	if stringer, ok := req.(fmt.Stringer); ok {
		return stringer.String()
	}
	return fmt.Sprintf("%+v", req)
}
