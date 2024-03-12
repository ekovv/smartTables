package domains

import "context"

type Service interface {
	ExecQuery(ctx context.Context, query string, user string) ([][]interface{}, error)
	Registration(ctx context.Context, user, password string) error
	Login(ctx context.Context, user, password string) error
	GetConnection(user, connect string)
	GetTables(ctx context.Context, user string) ([]string, error)
}
