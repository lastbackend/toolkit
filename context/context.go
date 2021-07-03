package context

import (
	"context"
	"github.com/lastbackend/engine/util/context/metadata"
)

func SetMetadata(ctx context.Context, k, v string) context.Context {
	return metadata.Set(ctx, k, v)
}

func GetMetadata(ctx context.Context, k string) (string, bool) {
	return metadata.Get(ctx, k)
}
