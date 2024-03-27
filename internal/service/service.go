package service

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"smartTables/config"
	"smartTables/internal/constants"
	"smartTables/internal/domains"
	"smartTables/internal/shema"
	"strings"
	"time"
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

func (s *Service) ExecQuery(ctx context.Context, query string, user string) ([][]string, error) {
	const op = "service.ExecQuery"
	var connectionString *sql.DB
	connection, ok := s.connections[user]
	for _, conn := range connection {
		if conn.Flag {
			connectionString = conn.Conn
		}
	}
	if !ok {
		s.logger.Info(fmt.Sprintf("%s : %v", op, fmt.Errorf("no connection")))
		return nil, fmt.Errorf("no connections")
	}

	if strings.Contains(query, "INSERT") || strings.Contains(query, "DELETE") || strings.Contains(query, "UPDATE") {
		err := ExecWithoutRes(ctx, query, connectionString)
		if err != nil {
			s.logger.Info(fmt.Sprintf("%s : %v", op, err))
			return nil, err
		}
		return nil, nil
	}

	res, err := ExecWithRes(ctx, query, connectionString)
	if err != nil {
		s.logger.Info(fmt.Sprintf("%s : %v", op, err))
		return nil, err
	}

	return res, nil
}
func ExecWithRes(ctx context.Context, query string, connectionString *sql.DB) ([][]string, error) {
	rows, err := connectionString.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	result := make([][]string, 0)

	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		strRow := make([]string, len(cols))
		for i, val := range columns {
			bytes, ok := val.([]byte)
			if ok {
				strRow[i] = string(bytes)
			} else {
				strRow[i] = fmt.Sprint(val)
			}
		}
		result = append(result, strRow)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func ExecWithoutRes(ctx context.Context, query string, connectionString *sql.DB) error {
	_, err := connectionString.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to do query: %w", err)
	}
	return nil
}

func (s *Service) GetConnection(ctx context.Context, user, typeDB, connect, dbName string) {
	const op = "service.GetConnection"
	c := shema.Connection{}
	c.TypeDB = typeDB
	c.Flag = true
	if dbName == "" {
		c.DBName = "DatabaseWithoutName"
	} else {
		c.DBName = dbName
	}
	var driver string
	switch {
	case typeDB == "postgresql":
		driver = "postgres"
	case typeDB == "mysql":
		driver = "mysql"
	}
	err := s.storage.SaveConnection(ctx, user, connect)
	if err != nil {
		s.logger.Info(fmt.Sprintf("%s : %v", op, err))
		return
	}
	db, err := sql.Open(driver, connect)
	if err != nil {
		s.logger.Info(fmt.Sprintf("%s : %v", op, err))
		return
	}
	c.Conn = db
	s.connections[user] = append(s.connections[user], c)
}

func (s *Service) GetConnectionWithFile(user, typeDB, dbName string, file *multipart.FileHeader) {
	const op = "service.GetConnection"
	userDir, err := createUserDir(user)
	if err != nil {
		return
	}

	fileRes, err := file.Open()
	if err != nil {
		return
	}
	defer fileRes.Close()

	dst, err := saveFile(userDir, file.Filename, fileRes)
	if err != nil {
		return
	}
	c := shema.Connection{}
	c.TypeDB = typeDB
	c.Flag = true
	if dbName == "" {
		c.DBName = "DatabaseWithoutName"
	} else {
		c.DBName = dbName
	}
	db, err := sql.Open("sqlite3", dst)
	if err != nil {
		s.logger.Info(fmt.Sprintf("%s : %v", op, err))
		return
	}
	c.Conn = db
	s.connections[user] = append(s.connections[user], c)
}
func createUserDir(username string) (string, error) {
	userDir := filepath.Join(".", username)
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		err = os.Mkdir(userDir, 0755)
		if err != nil {
			return "", err
		}
	}
	return userDir, nil
}

// Сохранение файла в директории пользователя
func saveFile(userDir string, filename string, file io.Reader) (string, error) {
	dst := filepath.Join(userDir, filename)
	dstFile, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, file)
	if err != nil {
		return "", err
	}

	return dst, nil
}

func (s *Service) Registration(ctx context.Context, user, password string) error {
	const op = "service.Registration"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	err = s.storage.Registration(ctx, user, hashedPassword)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return constants.ErrAlreadyExists
		} else {
			s.logger.Info(fmt.Sprintf("%s : %v", op, err))
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
	const op = "service.GetTables"
	var connectionString *sql.DB
	connection, ok := s.connections[user]
	typeDB := ""
	for _, conn := range connection {
		if conn.Flag {
			typeDB = conn.TypeDB
			connectionString = conn.Conn
		}
	}
	if !ok {
		return nil, fmt.Errorf("no connections")
	}
	res, err := GetAllTables(ctx, connectionString, typeDB)
	if err != nil {
		s.logger.Info(fmt.Sprintf("%s : %v", op, err))
		return nil, fmt.Errorf("can't get tables: %w", err)
	}
	return res, nil
}
func GetAllTables(ctx context.Context, connectionString *sql.DB, dbType string) ([]string, error) {
	var query string
	if dbType == "postgresql" {
		query = "SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema'"
	} else if dbType == "mysql" {
		query = "SHOW TABLES"
	} else if dbType == "sqlite" {
		query = "SELECT name FROM sqlite_master WHERE type='table'"
	} else {
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	rows, err := connectionString.QueryContext(ctx, query)
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

func (s *Service) QueryFromFile(ctx context.Context, file *multipart.FileHeader, user string) ([][]string, error) {
	const op = "service.QueryFromFile"
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
		s.logger.Info(fmt.Sprintf("%s : %v", op, err))
		return nil, err
	}
	if strings.Contains(string(fileBytes), "INSERT") || strings.Contains(string(fileBytes), "DELETE") || strings.Contains(string(fileBytes), "UPDATE") {
		err := ExecWithoutRes(ctx, string(fileBytes), connectionString)
		if err != nil {
			s.logger.Info(fmt.Sprintf("%s : %v", op, err))
			return nil, err
		}
		return nil, nil
	}

	res, err := ExecWithRes(ctx, string(fileBytes), connectionString)
	if err != nil {
		s.logger.Info(fmt.Sprintf("%s : %v", op, err))
		return nil, err
	}
	return res, nil
}

func (s *Service) Logout(user string) error {
	connection, ok := s.connections[user]
	if !ok {
		return nil
	}
	for i := range connection {
		if connection[i].Flag {
			connection[i].Flag = false
		}
	}

	return nil
}

func (s *Service) SaveQuery(ctx context.Context, query, user string) error {
	const op = "service.SaveQuery"
	t := time.Now()
	t.Format("2006-01-02 15:04:05")
	typeDB := ""
	dbName := ""
	connection, ok := s.connections[user]
	for _, conn := range connection {
		if conn.Flag {
			typeDB = conn.TypeDB
			dbName = conn.DBName
		}
	}
	if !ok {
		return fmt.Errorf("no connections")
	}
	err := s.storage.SaveQuery(ctx, user, typeDB, dbName, query, t)
	if err != nil {
		s.logger.Info(fmt.Sprintf("%s : %v", op, err))
		return fmt.Errorf("can't save query")
	}
	return nil
}

func (s *Service) GetHistory(ctx context.Context, user string) ([][]string, error) {
	const op = "service.GetHistory"
	connection, ok := s.connections[user]
	if !ok {
		return nil, fmt.Errorf("no connections")
	}
	dbName := ""
	for _, conn := range connection {
		if conn.Flag {
			dbName = conn.DBName
		}
	}
	res, err := s.storage.GetHistory(ctx, user, dbName)
	if err != nil {
		s.logger.Info(fmt.Sprintf("%s : %v", op, err))
		return nil, fmt.Errorf("can't get history: %w", err)
	}
	return res, nil
}

func (s *Service) Switch(user, typeDB string) error {
	const op = "service.Switch"
	connection, ok := s.connections[user]
	if !ok {
		return fmt.Errorf("no connections")
	}
	for i := range connection {
		if connection[i].Flag && connection[i].TypeDB == typeDB {
			connection[i].Flag = false
		}
	}

	return nil
}

func (s *Service) GetLastDB(ctx context.Context, user string) (map[string]string, error) {
	const op = "service.GetLastDB"
	m, err := s.storage.GetLastDB(ctx, user)
	if err != nil {
		s.logger.Info(fmt.Sprintf("%s : %v", op, err))
		return nil, fmt.Errorf("can't get last db: %w", err)
	}
	return m, nil
}
