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
	"time"
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

func (s *Storage) SaveConnection(ctx context.Context, user, connectionString string) error {
	sqlStatement := `
		UPDATE users SET connectionString = $1 WHERE login = $2`
	_, err := s.conn.ExecContext(ctx, sqlStatement, connectionString, user)
	if err != nil {
		return fmt.Errorf("unable to execute the query. %v", err)
	}
	return nil
}

func (s *Storage) GetLastDB(ctx context.Context, user string) (map[string]string, error) {
	result := make(map[string]string)

	query := `
        SELECT u.connectionString, h.dbName
        FROM users u
        JOIN history h ON u.login = h.login
        WHERE u.login = $1 AND h.time > NOW() - INTERVAL '10 days'
    `

	rows, err := s.conn.QueryContext(ctx, query, user)
	if err != nil {
		return nil, fmt.Errorf("unable to execute the query. %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var connectionString, dbName string
		if err := rows.Scan(&connectionString, &dbName); err != nil {
			return nil, fmt.Errorf("unable to scan the row. %v", err)
		}
		result[dbName] = connectionString
	}

	return result, nil
}

func (s *Storage) SaveQuery(ctx context.Context, user, typeDB, dbName, query string, time time.Time) error {
	sqlStatement := `
		INSERT INTO history (login, typeDB, dbName, time, query)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := s.conn.ExecContext(ctx, sqlStatement, user, typeDB, dbName, time, query)
	if err != nil {
		return fmt.Errorf("unable to execute the query. %v", err)
	}
	return nil
}

func (s *Storage) GetHistory(ctx context.Context, user, dbName string) ([][]string, error) {
	sqlStatement := `
		SELECT dbname, typedb, query, time
		FROM history 
		WHERE login = $1 AND dbName = $2`
	rows, err := s.conn.QueryContext(ctx, sqlStatement, user, dbName)
	if err != nil {
		return nil, fmt.Errorf("unable to execute the query. %v", err)
	}
	defer rows.Close()

	var history [][]string
	for rows.Next() {
		var dbname, typedb, query, time string
		err = rows.Scan(&dbname, &typedb, &query, &time)
		if err != nil {
			return nil, err
		}
		history = append(history, []string{dbname, typedb, query, time})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}
