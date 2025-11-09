package connection

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var mmUserInit sync.Once
var mmUserDb *PostgresMainDB

type PostgresMainDB struct {
	*pgxpool.Pool
}

func InjectMainDB() *PostgresMainDB {
	mmUserInit.Do(func() {
		s := setupMainUser()
		mmUserDb = s
	})
	return mmUserDb
}

func setupMainUser() *PostgresMainDB {
	ctx := context.Background()
	config, err := pgxpool.ParseConfig(os.Getenv("PGX_MAIN"))
	if err != nil {
		panic(err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		panic(err)
	}
	log.Println("DB CONNECTED")
	return &PostgresMainDB{pool}
}
