package domains

import "context"

type Service interface {
	ExecQuery(ctx context.Context, query string) ([][]interface{}, error)
}
