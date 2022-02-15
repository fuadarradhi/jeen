package jeen

import (
	"context"
	"database/sql"
	"time"
)

type Database struct {
	Context context.Context
	Conn    *sql.Conn
}

// Conn returns a single connection by either opening a new connection
// or returning an existing connection from the connection pool. Conn will
// block until either a connection is returned or ctx is canceled.
// Queries run on the same Conn will be run in the same database session.
//
// Every Conn must be returned to the database pool after use by
// calling Database.Close.
func conn(ctx context.Context, db *sql.DB) (*Database, error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}

	// only for ping, timeout context 3 second
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err = conn.PingContext(pingCtx)
	if err != nil {
		return nil, err
	}

	return &Database{
		Context: ctx,
		Conn:    conn,
	}, nil
}

// Close returns the connection to the connection pool.
// All operations after a Close will return with ErrConnDone.
// Close is safe to call concurrently with other operations and will
// block until all other operations finish. It may be useful to first
// cancel any used context and then call close directly after.
func (d *Database) Close() {
	if d.Conn != nil {
		d.Conn.Close()
	}
}
