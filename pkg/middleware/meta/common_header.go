package meta

import "context"

var (
	userId     = "x-md-userid"
	orgId      = "x-md-orgid"
	datasource = "x-md-datasource"
)

type commonHeaderKey struct{}

type CommonHeader struct {
	OrgId      int64
	UserId     int64
	Datasource string
}

func GetCommonHeaderFromCtx(ctx context.Context) *CommonHeader {
	return ctx.Value(commonHeaderKey{}).(*CommonHeader)
}

func SetCommonHeader(ctx context.Context, header *CommonHeader) context.Context {
	return context.WithValue(ctx, commonHeaderKey{}, header)
}
