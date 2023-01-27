package context

import (
	"github.com/lastbackend/toolkit/pkg/context/metadata"

	"context"
)

func SetMetadata(ctx context.Context, k, v string) context.Context {
	return metadata.SetToContext(ctx, k, v)
}

func GetMetadata(ctx context.Context, k string) (string, bool) {
	return metadata.FindInContext(ctx, k)
}
