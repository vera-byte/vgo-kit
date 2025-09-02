package vgokit

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// GetMetadataValue 获取上下文的元数据
func GetMetadataValue(ctx context.Context, field string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get(field)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}
