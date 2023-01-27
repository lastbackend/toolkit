package metadata

import (
	"context"
)

const emptyString = ""

type MD map[string]string
type mdKey struct{}

func (m MD) Get(key string) (string, bool) {
	val, ok := m[key]
	if ok {
		return val, ok
	}
	val, ok = m[key]
	return val, ok
}

func (m MD) Set(key, val string) {
	m[key] = val
}

func (m MD) Del(key string) {
	delete(m, key)
}

func NewContext(ctx context.Context, md MD) context.Context {
	return context.WithValue(ctx, mdKey{}, md)
}

func DelFromContext(ctx context.Context, k string) context.Context {
	return SetToContext(ctx, k, emptyString)
}

func SetToContext(ctx context.Context, k, v string) context.Context {
	md, ok := LoadFromContext(ctx)
	if !ok {
		md = make(MD, 0)
	}
	if len(v) != 0 {
		md[k] = v
	} else {
		delete(md, k)
	}
	return context.WithValue(ctx, mdKey{}, md)
}

func FindInContext(ctx context.Context, key string) (string, bool) {
	md, ok := LoadFromContext(ctx)
	if !ok {
		return emptyString, ok
	}
	val, ok := md[key]
	return val, ok
}

func LoadFromContext(ctx context.Context) (MD, bool) {
	md, ok := ctx.Value(mdKey{}).(MD)
	if !ok {
		return nil, ok
	}
	newMD := make(MD, len(md))
	for k, v := range md {
		newMD[k] = v
	}
	return newMD, ok
}

func MergeContext(ctx context.Context, metadata MD, overwrite bool) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	md, ok := ctx.Value(mdKey{}).(MD)
	if !ok {
		return ctx
	}
	cp := make(MD, len(md))
	for k, v := range md {
		cp[k] = v
	}
	for k, v := range metadata {
		if _, ok := cp[k]; ok && !overwrite {
			continue
		}
		if len(v) != 0 {
			cp[k] = v
		} else {
			delete(cp, k)
		}
	}
	return context.WithValue(ctx, mdKey{}, cp)
}
