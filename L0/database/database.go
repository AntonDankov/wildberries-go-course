package database

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const MigrationFolderPath = "migration/"

type Database struct {
	Pool *pgxpool.Pool
}

func New() *Database {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Not found .env file ", err)
	}

	connection := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DATABASE_USER"),
		os.Getenv("DATABASE_PASSWORD"),
		os.Getenv("DATABASE_HOST"),
		os.Getenv("DATABASE_PORT"),
		os.Getenv("DATABASE_NAME"),
		os.Getenv("DATABASE_SSL_MODE"),
	)

	config, err := pgxpool.ParseConfig(connection)
	if err != nil {
		log.Fatal("Failed to parse database config: ", err)
	}

	config.MaxConns = 42
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatal("Failed to create database connection pool: ", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("Failed to ping the database: ", err)
	}

	return &Database{Pool: pool}
}

func (db *Database) RunMigration(folderPath string) error {
	ctx := context.Background()

	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return fmt.Errorf("failed to read migration folder %s: %w", folderPath, err)
	}

	var filePaths []string
	for _, entry := range entries {
		slog.Debug("Entry migration file: ", "name", entry)
		if !entry.IsDir() {
			filePaths = append(filePaths, filepath.Join(folderPath, entry.Name()))
		}
	}

	sort.Strings(filePaths)
	for _, filepath := range filePaths {

		sql, err := os.ReadFile(filepath)
		if err != nil {
			return fmt.Errorf("failed to read migration file: %v", err)
		}

		_, err = db.Pool.Exec(ctx, string(sql))
		if err != nil {
			return fmt.Errorf("failed to run migration file: %v", err)
		}

		slog.Info("Executed migration file: ", "filepath", filepath)
	}

	return nil
}
