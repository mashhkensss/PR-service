package service

import "context"

func RunInTx[T any](ctx context.Context, tx TxRunner, fn func(ctx context.Context) (T, error)) (T, error) {
	var zero T
	if tx == nil {
		return fn(ctx)
	}
	var (
		result T
		err    error
	)
	runErr := tx.WithinTx(ctx, func(ctx context.Context) error {
		result, err = fn(ctx)
		return err
	})
	if runErr != nil {
		return zero, runErr
	}
	return result, nil
}

func ExecInTx(ctx context.Context, tx TxRunner, fn func(ctx context.Context) error) error {
	if tx == nil {
		return fn(ctx)
	}
	return tx.WithinTx(ctx, fn)
}
