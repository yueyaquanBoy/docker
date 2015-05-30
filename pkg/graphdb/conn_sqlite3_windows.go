// +build cgo,windows

package graphdb

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3" // registers sqlite
)

func NewSqliteConn(root string) (*Database, error) {
	conn, err := sql.Open("sqlite3", root)
	if err != nil {
		return nil, err
	}

	return NewDatabase(conn)
}
