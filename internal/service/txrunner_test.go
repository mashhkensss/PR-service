package service

import (
	"context"
	"errors"
	"testing"
)

type testTxRunner struct {
	err error
}

func (r testTxRunner) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if r.err != nil {
		return r.err
	}
	return fn(ctx)
}

func TestRunInTx_WithTx(t *testing.T) {
	tx := testTxRunner{}
	got, err := RunInTx(context.Background(), tx, func(ctx context.Context) (string, error) {
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ok" {
		t.Fatalf("unexpected result: %s", got)
	}
}

func TestRunInTx_NoTx(t *testing.T) {
	got, err := RunInTx(context.Background(), nil, func(ctx context.Context) (int, error) {
		return 42, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
}

func TestRunInTx_Error(t *testing.T) {
	expected := errors.New("boom")
	tx := testTxRunner{}
	_, err := RunInTx(context.Background(), tx, func(ctx context.Context) (int, error) {
		return 0, expected
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestExecInTx_WithTx(t *testing.T) {
	tx := testTxRunner{}
	if err := ExecInTx(context.Background(), tx, func(ctx context.Context) error {
		return nil
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecInTx_NoTx(t *testing.T) {
	if err := ExecInTx(context.Background(), nil, func(ctx context.Context) error {
		return nil
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecInTx_Error(t *testing.T) {
	expected := errors.New("boom")
	if err := ExecInTx(context.Background(), nil, func(ctx context.Context) error {
		return expected
	}); !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}
