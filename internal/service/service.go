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
	"smartTables/internal/shema"
	"strings"
)

type Service struct {
	storage     domains.Storage
	config      config.Config
	logger      *zap.Logger
	connections map[string]shema.Connections
}

func NewService(storage domains.Storage, config config.Config) *Service {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil
	}
	return &Service{storage: storage, logger: logger, config: config, connections: make(map[string]shema.Connections)}
}

func (s *Service) ExecQuery(ctx context.Context, query string) ([][]interface{}, error) {
	res, err := s.storage.ExecWithRes(ctx, query)
	if err != nil {
		return nil, err
	}
	return res, nil

}

func (s *Service) GetConnection(cook, user, password, connect string) {
	conn := shema.Connections{}
	conn.Login = user
	conn.Password = password
	conn.ConnectionDB = connect
	s.connections[cook] = conn
}

func (s *Service) Registration(ctx context.Context, user, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	err = s.storage.Registration(ctx, user, hashedPassword)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("login already exists")
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
			return fmt.Errorf("not checked")
		}
	}
	err = bcrypt.CompareHashAndPassword(pass, []byte(password))
	if err != nil {
		return constants.ErrInvalidData
	}
	return nil
}
