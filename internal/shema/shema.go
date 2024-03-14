package shema

import "database/sql"

type Connection struct {
	TypeDB string
	DBName string
	Conn   *sql.DB
	Flag   bool
}
