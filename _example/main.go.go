package main

import (
	"database/sql"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/fuadarradhi/jeen"
	"github.com/gomodule/redigo/redis"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {

	serv := jeen.InitServer(&jeen.Config{
		Driver: &jeen.Driver{
			// Database driver
			Database: func() *sql.DB {
				db, err := sql.Open("pgx", dsn())
				if err != nil {
					panic(err)
				}
				return db
			}(),
			// SCS Session store
			Session: func() scs.Store {
				pool := &redis.Pool{
					MaxIdle: 10,
					Dial: func() (redis.Conn, error) {
						return redis.Dial("tcp", "127.0.0.1:6379")
					},
				}
				return redisstore.New(pool)
			}(),
		},
		// Default value
		Default: &jeen.Default{
			WithDatabase: true,
			WithTimeout:  7 * time.Second,
		},
	})
	defer serv.Close()

	serv.Get("/", func(res *jeen.Resource) {
		res.Writer.Write([]byte("TES"))
	})

	serv.ListenAndServe(":8000")
}
