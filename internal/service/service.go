package service

import (
	"context"
	"go.uber.org/zap"
	"smartTables/config"
	"smartTables/internal/domains"
)

type Service struct {
	storage domains.Storage
	config  config.Config
	logger  *zap.Logger
}

func NewService(storage domains.Storage, config config.Config) *Service {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil
	}
	return &Service{storage: storage, logger: logger, config: config}
}

func (s *Service) ExecQuery(ctx context.Context, query string) ([][]interface{}, error) {
	res, err := s.storage.ExecWithRes(ctx, query)
	if err != nil {
		return nil, err
	}
	return res, nil

}
