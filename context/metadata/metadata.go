package metadata

import (
	"context"
	"strings"
)

type metadataKey struct{}

type Metadata map[string]string

func (md Metadata) Get(key string) (string, bool) {
	val, ok := md[key]
	if ok {
		return val, ok
	}
	val, ok = md[strings.Title(key)]
	return val, ok
}

func (md Metadata) Set(key, val string) {
	md[key] = val
}

func (md Metadata) Delete(key string) {
	delete(md, key)
	delete(md, strings.Title(key))
}

func Copy(md Metadata) Metadata {
	cmd := make(Metadata, len(md))
	for k, v := range md {
		cmd[k] = v
	}
	return cmd
}

func Delete(ctx context.Context, k string) context.Context {
	return Set(ctx, k, "")
}

func Set(ctx context.Context, k, v string) context.Context {
	md, ok := FromContext(ctx)
	if !ok {
		md = make(Metadata)
	}
	if v == "" {
		delete(md, k)
	} else {
		md[k] = v
	}
	return context.WithValue(ctx, metadataKey{}, md)
}

func Get(ctx context.Context, key string) (string, bool) {
	md, ok := FromContext(ctx)
	if !ok {
		return "", ok
	}
	val, ok := md[key]
	if ok {
		return val, ok
	}
	val, ok = md[strings.Title(key)]
	return val, ok
}

func FromContext(ctx context.Context) (Metadata, bool) {
	md, ok := ctx.Value(metadataKey{}).(Metadata)
	if !ok {
		return nil, ok
	}
	newMD := make(Metadata, len(md))
	for k, v := range md {
		newMD[strings.Title(k)] = v
	}
	return newMD, ok
}

func NewContext(ctx context.Context, md Metadata) context.Context {
	return context.WithValue(ctx, metadataKey{}, md)
}

func MergeContext(ctx context.Context, patchMd Metadata, overwrite bool) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	md, _ := ctx.Value(metadataKey{}).(Metadata)
	cmd := make(Metadata, len(md))
	for k, v := range md {
		cmd[k] = v
	}
	for k, v := range patchMd {
		if _, ok := cmd[k]; ok && !overwrite {
		} else if v != "" {
			cmd[k] = v
		} else {
			delete(cmd, k)
		}
	}
	return context.WithValue(ctx, metadataKey{}, cmd)
}
