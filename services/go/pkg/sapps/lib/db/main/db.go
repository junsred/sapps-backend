package maindb

import (
	"context"
	"sapps/lib/connection"
)

type MainDB struct {
	*connection.PostgresMainDB
}

func InjectMainDB(db *connection.PostgresMainDB) *MainDB {
	return &MainDB{
		db,
	}
}

func (q *MainDB) InsertCost(ctx context.Context, userID string, cost float64, reason string, taskID string, ipAddress string) error {
	query := `INSERT INTO costs (user_id, price, reason, task_id, ip_address) VALUES ($1, $2, $3, $4, $5);`
	_, err := q.Exec(ctx, query, userID, cost, reason, taskID, ipAddress)
	return err
}
