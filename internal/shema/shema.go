package shema

import "database/sql"

type Connection struct {
	TypeDB string
	Conn   *sql.DB
	Flag   bool
}
