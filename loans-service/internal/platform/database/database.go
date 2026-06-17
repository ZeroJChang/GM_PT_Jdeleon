package database

import (
"context"
"database/sql"
"os"
"time"

_ "github.com/jackc/pgx/v5/stdlib"
)

const defaultDatabaseURL = "postgres://postgres:postgres@localhost:5434/loans_db?sslmode=disable"

func Open(ctx context.Context) (*sql.DB, error) {
databaseURL := os.Getenv("DATABASE_URL")
if databaseURL == "" {
databaseURL = defaultDatabaseURL
}

db, err := sql.Open("pgx", databaseURL)
if err != nil {
return nil, err
}

db.SetMaxOpenConns(10)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(30 * time.Minute)

if err := db.PingContext(ctx); err != nil {
return nil, err
}

return db, nil
}

func RunMigrations(ctx context.Context, db *sql.DB) error {
migration, err := os.ReadFile("migrations/001_init.sql")
if err != nil {
return err
}

_, err = db.ExecContext(ctx, string(migration))
return err
}
