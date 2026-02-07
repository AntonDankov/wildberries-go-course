package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/joho/godotenv"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

const MigrationFolderPath = "migration/"

type Database struct {
	*dbpg.DB
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

	opts := &dbpg.Options{
		MaxOpenConns:    42,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Hour,
	}

	slaveDsns := []string{}

	db, err := dbpg.New(connection, slaveDsns, opts)
	if err != nil {
		log.Fatal("Failed to create database connection pool: ", err)
	}

	if err := db.Master.Ping(); err != nil {
		log.Fatal("Failed to ping the database: ", err)
	}

	return &Database{db}
}

func (db *Database) RunMigration(folderPath string) error {
	ctx := context.Background()
	zlog.Logger.Info().Msgf("Executed migration file: %s", "test")
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return fmt.Errorf("failed to read migration folder %s: %w", folderPath, err)
	}

	var filePaths []string
	for _, entry := range entries {
		zlog.Logger.Info().Msgf("Entry migration file: %s", entry)
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

		_, err = db.Master.ExecContext(ctx, string(sql))
		if err != nil {
			return fmt.Errorf("failed to run migration file: %v", err)
		}
		zlog.Logger.Info().Msgf("Executed migration file: %s", filepath)
	}

	return nil
}
