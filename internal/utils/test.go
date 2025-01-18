package utils

import (
	"context"
	"sync"
)

type contextKey string

const waitGroupKey contextKey = "waitGroup"

func AddWaitGroupToContext(ctx context.Context, wg *sync.WaitGroup) context.Context {
	return context.WithValue(ctx, waitGroupKey, wg)
}

func GetWaitGroupFromContext(ctx context.Context) (*sync.WaitGroup, bool) {
	wg, ok := ctx.Value(waitGroupKey).(*sync.WaitGroup)
	return wg, ok
}
