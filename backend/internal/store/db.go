package store

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

func ConnectDB(host, port, user, password, dbname string) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	var pingErr error
	for i := 1; i <= 10; i++ {
		pingErr = db.Ping()
		if pingErr == nil {
			slog.Info("Database connected", "attempt", i)
			db.SetMaxOpenConns(25)
			db.SetMaxIdleConns(25)
			db.SetConnMaxLifetime(5 * time.Minute)
			return db, nil
		}
		slog.Warn("Database not ready, retrying", "attempt", i, "max", 10, "err", pingErr)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to ping db after 10 attempts: %w", pingErr)
}
