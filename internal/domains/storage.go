package domains

import "context"

type Storage interface {
	ExecWithRes(ctx context.Context, query string) ([][]interface{}, error)
	Registration(ctx context.Context, user string, password []byte) error
	Login(ctx context.Context, user string) ([]byte, error)
}
