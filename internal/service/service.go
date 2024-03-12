package service

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"log"
	"smartTables/config"
	"smartTables/internal/constants"
	"smartTables/internal/domains"
	"strings"
)

type Service struct {
	storage     domains.Storage
	config      config.Config
	logger      *zap.Logger
	connections map[string]string
}

func NewService(storage domains.Storage, config config.Config) *Service {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil
	}
	return &Service{storage: storage, logger: logger, config: config, connections: make(map[string]string)}
}

func (s *Service) ExecQuery(ctx context.Context, query string, user string) ([][]interface{}, error) {
	connectionString, ok := s.connections[user]
	if !ok {
		return nil, fmt.Errorf("no connections")
	}

	if strings.Contains(query, "INSERT") || strings.Contains(query, "DELETE") || strings.Contains(query, "UPDATE") {
		err := s.storage.ExecWithoutRes(ctx, query, connectionString)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	res, err := s.storage.ExecWithRes(ctx, query, connectionString)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Service) GetConnection(user, connect string) {
	s.connections[user] = connect
}

func (s *Service) Registration(ctx context.Context, user, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	err = s.storage.Registration(ctx, user, hashedPassword)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return constants.ErrAlreadyExists
		} else {
			return fmt.Errorf("not saved")
		}
	}

	return nil
}

func (s *Service) Login(ctx context.Context, user, password string) error {
	pass, err := s.storage.Login(ctx, user)
	if err != nil {
		if strings.Contains(err.Error(), "user not registered") {
			return constants.ErrInvalidData
		} else {
			return constants.ErrInvalidData
		}
	}

	err = bcrypt.CompareHashAndPassword(pass, []byte(password))
	if err != nil {
		return constants.ErrInvalidData
	}

	return nil
}

func (s *Service) GetTables(ctx context.Context, user string) ([]string, error) {
	connectionString, ok := s.connections[user]
	if !ok {
		return nil, fmt.Errorf("no connections")
	}
	res, err := s.storage.GetAllTables(ctx, connectionString)
	if err != nil {
		return nil, fmt.Errorf("can't get tables: %w", err)
	}
	return res, nil
}
