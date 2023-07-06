package meta

import (
	"context"
	"strings"

	"github.com/spf13/cast"

	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// Option is metadata option.
type Option func(*options)

type options struct {
	prefix []string
	md     metadata.Metadata
}

func (o *options) hasPrefix(key string) bool {
	k := strings.ToLower(key)
	for _, prefix := range o.prefix {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}
	return false
}

// Server is middleware server-side metadata.
func Server(opts ...Option) middleware.Middleware {
	options := &options{
		prefix: []string{"x-md-", "pm-"}, // x-md-global-, x-md-local
	}
	for _, o := range opts {
		o(options)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				md := options.md.Clone()
				header := tr.RequestHeader()
				commonHeader := &CommonHeader{}
				for _, k := range header.Keys() {
					if options.hasPrefix(k) {
						md.Set(k, header.Get(k))
					}
					switch strings.ToLower(k) {
					case userId:
						commonHeader.UserId = cast.ToInt64(header.Get(k))
					case orgId:
						commonHeader.OrgId = cast.ToInt64(header.Get(k))
					case datasource:
						commonHeader.Datasource = header.Get(k)
					}
				}
				ctx = metadata.NewServerContext(ctx, md)
				ctx = context.WithValue(ctx, commonHeaderKey{}, commonHeader)
			}
			return handler(ctx, req)
		}
	}
}

// Client is middleware client-side metadata.
func Client(opts ...Option) middleware.Middleware {
	options := &options{
		prefix: []string{"x-md-", "pm-"},
	}
	for _, o := range opts {
		o(options)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if tr, ok := transport.FromClientContext(ctx); ok {
				header := tr.RequestHeader()
				// x-md-local-
				for k, v := range options.md {
					header.Set(k, v)
				}
				if md, ok := metadata.FromClientContext(ctx); ok {
					for k, v := range md {
						header.Set(k, v)
					}
				}
				// x-md-global-
				if md, ok := metadata.FromServerContext(ctx); ok {
					for k, v := range md {
						if options.hasPrefix(k) {
							header.Set(k, v)
						}
					}
				}
			}
			return handler(ctx, req)
		}
	}
}
