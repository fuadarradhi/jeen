package jeen

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/georgysavva/scany/dbscan"
	"github.com/georgysavva/scany/sqlscan"
)

type Database struct {
	// request context
	context context.Context

	// scanny db scan
	scan *dbscan.API

	// database/sql DB pool, can be used by other packages that require a *sql.DB
	DB *sql.DB

	// database/sql Conn, can be used by other packages that require
	// a single connection from *sql.DB
	Conn *sql.Conn
}

type SqlQuery struct {
	// request context
	context context.Context

	// scanny scan
	scan *dbscan.API

	// conn from sql.DB
	conn *sql.Conn

	// save query before get result or scan to struct
	query string

	// save args before get result or scan to struct
	args []interface{}
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

	scan, err := sqlscan.NewDBScanAPI(dbscan.WithStructTagKey("jeen"))
	if err != nil {
		log.Fatal("error when create scanny Api")
	}

	return &Database{
		context: ctx,
		scan:    scan,
		Conn:    conn,
		DB:      db,
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

// buildquery will convert named queries to positional queries, so they can
// be executed directly by the database/sql without changing the working
// system of sanitaze sql injection
func (d *Database) BuildQuery(namedQuery string, namedArgs ...Map) (string, []interface{}, error) {
	var query bytes.Buffer
	var named bytes.Buffer
	var namedIndex int = 1
	var char rune
	var width int
	var args []interface{}
	var field string

	_args := Map{}
	for _, arg := range namedArgs {
		for k, v := range arg {
			_args[k] = v
		}
	}

	for pos := 0; pos < len(namedQuery); {
		char, width = utf8.DecodeRuneInString(namedQuery[pos:])
		pos += width

		if char == ':' {
			named.Reset()
			for {
				char, width = utf8.DecodeRuneInString(namedQuery[pos:])
				pos += width

				if unicode.IsLetter(char) || unicode.IsDigit(char) ||
					char == '_' || char == '.' {
					named.WriteRune(char)
				} else {
					break
				}
			}
			if char == ':' {
				query.WriteRune(char)
			} else {

				field = named.String()
				if val, ok := _args[field]; ok {
					args = append(args, val)
				} else {
					return query.String(), args, fmt.Errorf(`field '%s' is not defined`, field)
				}

				// TODO: for all drivers
				query.WriteRune('$')
				query.WriteString(strconv.Itoa(namedIndex))

				namedIndex++
			}
		}

		if char <= unicode.MaxASCII {
			query.WriteRune(char)
		}
	}

	return query.String(), args, nil
}

// Query will only store the query string and arguments without executing it.
// see Result, Row and Exec for more information.
func (d *Database) Query(query string, args ...Map) *SqlQuery {
	qry, arg, err := d.BuildQuery(query, args...)
	if err != nil {
		log.Fatal(err)
	}

	return &SqlQuery{
		context: d.context,
		conn:    d.Conn,
		scan:    d.scan,
		query:   qry,
		args:    arg,
	}
}

// Result will return all rows from the query
func (q *SqlQuery) Result(dest interface{}) error {
	rows, err := q.conn.QueryContext(q.context, q.query, q.args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	err = q.scan.ScanAll(dest, rows)
	if err != nil {
		return err
	}

	return nil
}

// Row will return only one row from the query
func (q *SqlQuery) Row(dest interface{}) error {
	rows, err := q.conn.QueryContext(q.context, q.query, q.args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	err = q.scan.ScanOne(dest, rows)
	if err != nil {
		return err
	}
	return nil
}

// Exec execute query
func (q *SqlQuery) Exec() (sql.Result, error) {
	return q.conn.ExecContext(q.context, q.query, q.args...)
}
