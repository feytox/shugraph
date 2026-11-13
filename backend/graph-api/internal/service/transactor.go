package service

import "context"

type Transactor[R any] func(ctx context.Context, fn func(R) error) error
