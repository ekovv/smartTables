package service

import (
	"context"
	"database/sql"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"mime/multipart"
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
	connections map[string][]shema.Connection
}

func NewService(storage domains.Storage, config config.Config) *Service {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil
	}
	return &Service{storage: storage, logger: logger, config: config, connections: make(map[string][]shema.Connection)}
}

func (s *Service) ExecQuery(ctx context.Context, query string, user string) ([][]interface{}, error) {
	var connectionString *sql.DB
	connection, ok := s.connections[user]
	for _, conn := range connection {
		if conn.Flag {
			connectionString = conn.Conn
		}
	}
	if !ok {
		return nil, fmt.Errorf("no connections")
	}

	if strings.Contains(query, "INSERT") || strings.Contains(query, "DELETE") || strings.Contains(query, "UPDATE") {
		err := ExecWithoutRes(ctx, query, connectionString)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	res, err := ExecWithRes(ctx, query, connectionString)
	if err != nil {
		return nil, err
	}

	return res, nil
}
func ExecWithRes(ctx context.Context, query string, connectionString *sql.DB) ([][]interface{}, error) {
	stmt, err := connectionString.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	colNames := make([]interface{}, len(cols))
	for i, v := range cols {
		colNames[i] = v
	}
	result := [][]interface{}{colNames}

	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}
		result = append(result, columns)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
func ExecWithoutRes(ctx context.Context, query string, connectionString *sql.DB) error {
	stmt, err := connectionString.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to do query: %w", err)
	}
	return nil
}

func (s *Service) GetConnection(user, typeDB, connect string) {
	c := shema.Connection{}
	c.TypeDB = typeDB
	c.Flag = true
	db, err := sql.Open("postgres", connect)
	if err != nil {
		return
	}
	c.Conn = db
	s.connections[user] = append(s.connections[user], c)
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
	var connectionString *sql.DB
	connection, ok := s.connections[user]
	for _, conn := range connection {
		if conn.Flag {
			connectionString = conn.Conn
		}
	}
	if !ok {
		return nil, fmt.Errorf("no connections")
	}
	res, err := GetAllTables(ctx, connectionString)
	if err != nil {
		return nil, fmt.Errorf("can't get tables: %w", err)
	}
	return res, nil
}
func GetAllTables(ctx context.Context, connectionString *sql.DB) ([]string, error) {
	rows, err := connectionString.QueryContext(ctx, "SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

func (s *Service) QueryFromFile(ctx context.Context, file *multipart.FileHeader, user string) ([][]interface{}, error) {
	if file == nil {
		return nil, fmt.Errorf("missing file")
	}
	var connectionString *sql.DB
	connection, ok := s.connections[user]
	for _, conn := range connection {
		if conn.Flag {
			connectionString = conn.Conn
		}
	}
	if !ok {
		return nil, fmt.Errorf("no connections")
	}
	f, err := file.Open()
	if err != nil {
		return nil, err
	}

	fileBytes, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if strings.Contains(string(fileBytes), "INSERT") || strings.Contains(string(fileBytes), "DELETE") || strings.Contains(string(fileBytes), "UPDATE") {
		err := ExecWithoutRes(ctx, string(fileBytes), connectionString)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	res, err := ExecWithRes(ctx, string(fileBytes), connectionString)
	if err != nil {
		return nil, err
	}
	return res, nil
}
