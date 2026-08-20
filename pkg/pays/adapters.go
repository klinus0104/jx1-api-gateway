package pays

import (
	"context"
	"time"
)

// PaySysClient and RelayClient are intentionally typed boundaries. Concrete
// protocol clients can be added without exposing raw opcodes to HTTP handlers.
type PaySysClient interface {
	GetAccount(ctx context.Context, account string) (AccountSnapshot, error)
	ResetPassword(ctx context.Context, account, password string) error
}

type RelayClient interface {
	GetSession(ctx context.Context, account string) (SessionSnapshot, error)
	Kick(ctx context.Context, account string) error
	Freeze(ctx context.Context, account string) error
	Unfreeze(ctx context.Context, account string) error
}

type AccountSnapshot struct {
	Name     string
	ClientID int64
	UserIP   int64
	Online   bool
}
type SessionSnapshot struct {
	Account  string
	ClientID int64
	UserIP   int64
	Online   bool
}

func withTimeout(ctx context.Context, timeout time.Duration, fn func(context.Context) error) error {
	child, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return fn(child)
}
