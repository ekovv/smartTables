package domains

import (
	"context"
	"mime/multipart"
)

type Service interface {
	ExecQuery(ctx context.Context, query string, user string) ([][]interface{}, error)
	Registration(ctx context.Context, user, password string) error
	Login(ctx context.Context, user, password string) error
	GetConnection(user, typeDB, connect string)
	GetTables(ctx context.Context, user string) ([]string, error)
	QueryFromFile(ctx context.Context, file *multipart.FileHeader, user string) ([][]interface{}, error)
	LogoutConnection(user, db string) error
}
