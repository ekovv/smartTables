package domains

import "context"

type Service interface {
	ExecQuery(ctx context.Context, query string) ([][]interface{}, error)
	Registration(ctx context.Context, user, password string) error
	Login(ctx context.Context, user, password string) error
}
