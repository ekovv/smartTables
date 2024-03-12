package domains

import "context"

type Storage interface {
	ExecWithRes(ctx context.Context, query string, connectionString string) ([][]interface{}, error)
	ExecWithoutRes(ctx context.Context, query, connectionString string) error
	Registration(ctx context.Context, user string, password []byte) error
	Login(ctx context.Context, user string) ([]byte, error)
	GetAllTables(ctx context.Context, connectionString string) ([]string, error)
}
