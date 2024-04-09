package domains

import (
	"context"
	"mime/multipart"
)

type Service interface {
	ExecQuery(ctx context.Context, query string, user string) ([][]string, error)
	Registration(ctx context.Context, user, password string) error
	Login(ctx context.Context, user, password string) error
	GetConnection(ctx context.Context, user, typeDB, connect, dbName string)
	GetConnectionWithFile(user, typeDB, dbName string, file *multipart.FileHeader)
	GetConnectionFromBtn(ctx context.Context, user, connect, dbName string) (string, error)
	GetTables(ctx context.Context, user string) ([]string, error)
	QueryFromFile(ctx context.Context, file *multipart.FileHeader, user string) ([][]string, error)
	Logout(user string) error
	SaveQuery(ctx context.Context, query, user string) error
	GetHistory(ctx context.Context, user string) ([][]string, error)
	Switch(user, typeDB string) error
	GetLastDB(ctx context.Context, user string) (map[string]string, error)
}
