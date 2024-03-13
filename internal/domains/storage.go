package domains

import "context"

type Storage interface {
	Registration(ctx context.Context, user string, password []byte) error
	Login(ctx context.Context, user string) ([]byte, error)
}
