package domains

import "context"

type Storage interface {
	ExecWithRes(ctx context.Context, query string) ([][]interface{}, error)
}
