package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"smartTables/config"
	"strings"
)

type Storage struct {
	conn *sql.DB
}

func NewPostgresDBStorage(config config.Config) (*Storage, error) {
	db, err := sql.Open("postgres", config.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db %w", err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate driver, %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"smartTables", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("failed to do migrate %w", err)
	}
	s := &Storage{
		conn: db,
	}

	return s, s.CheckConnection()
}

func (s *Storage) CheckConnection() error {
	if err := s.conn.Ping(); err != nil {
		return fmt.Errorf("failed to connect to db %w", err)
	}

	return nil
}

func (s *Storage) Close() error {
	return s.conn.Close()
}

func (s *Storage) Registration(ctx context.Context, user string, password []byte) error {
	_, err := s.conn.ExecContext(ctx, "INSERT INTO users (login, password) VALUES ($1, $2)", user, password)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("unique constraint")
		} else {
			return fmt.Errorf("not saved in database: %w", err)
		}
	}
	return nil
}

func (s *Storage) Login(ctx context.Context, user string) ([]byte, error) {
	var dbPassword []byte
	err := s.conn.QueryRowContext(ctx, "SELECT password FROM users WHERE login = $1", user).Scan(&dbPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not registered")
		}
		return nil, fmt.Errorf("failed to check: %w", err)
	}
	return dbPassword, nil
}
