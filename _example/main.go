package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/fuadarradhi/jeen"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	serv := jeen.InitServer(&jeen.Config{
		Driver: &jeen.Driver{

			// Database driver
			Database: func() *sql.DB {

				config, err := pgx.ParseConfig(os.Getenv("DSN"))
				if err != nil {
					panic(err)
				}
				config.PreferSimpleProtocol = true
				db := stdlib.OpenDB(*config)
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
			WithDatabase: false,
			WithTimeout:  7 * time.Second,
		},
	})
	defer serv.Close()

	serv.Get("/", func(res *jeen.Resource) {

		res.Html.Success()

	}, jeen.WithDatabase(true))

	serv.ListenAndServe(":8000")
}
