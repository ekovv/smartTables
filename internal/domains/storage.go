package domains

import (
	"context"
	"time"
)

type Storage interface {
	Registration(ctx context.Context, user string, password []byte) error
	Login(ctx context.Context, user string) ([]byte, error)
	SaveQuery(ctx context.Context, user, typeDB, dbName, query string, time time.Time) error
	GetHistory(ctx context.Context, user, dbName string) ([][]string, error)
	SaveConnection(ctx context.Context, user, connectionString string) error
	GetLastDB(ctx context.Context, user string) (map[string]string, error)
}
