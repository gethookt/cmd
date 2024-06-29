package trace

import (
	"context"
	"reflect"
)

type key struct{ string }

func makeKey[T any]() key {
	t := reflect.TypeOf((*T)(nil))
	return key{t.Elem().String()}
}

func with[T any](ctx context.Context, t *T) context.Context {
	return context.WithValue(ctx, makeKey[T](), t)
}

func from[T any](ctx context.Context) *T {
	t, _ := ctx.Value(makeKey[T]()).(*T)
	return t
}
