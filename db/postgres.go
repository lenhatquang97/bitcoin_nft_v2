package db

import (
	"bitcoin_nft_v2/db/sqlc"
	"database/sql"
	"fmt"
	postgres_migrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
)

const (
	dsnTemplate = "postgres://%v:%v@%v:%d/%v?sslmode=%v"
)

// PostgresConfig holds the postgres database configuration.
type PostgresConfig struct {
	SkipMigrations     bool   `long:"skipmigrations" description:"Skip applying migrations on startup."`
	Host               string `long:"host" description:"Database server hostname."`
	Port               int    `long:"port" description:"Database server port."`
	User               string `long:"user" description:"Database user."`
	Password           string `long:"password" description:"Database user's password."`
	DBName             string `long:"dbname" description:"Database name to use."`
	MaxOpenConnections int32  `long:"maxconnections" description:"Max open connections to keep alive to the database server."`
	RequireSSL         bool   `long:"requiressl" description:"Whether to require using SSL (mode: require) when connecting to the server."`
}

// DSN returns the dns to connect to the database.
func (s *PostgresConfig) DSN(hidePassword bool) string {
	var sslMode = "disable"
	if s.RequireSSL {
		sslMode = "require"
	}

	password := s.Password
	if hidePassword {
		// Placeholder used for logging the DSN safely.
		password = "****"
	}

	return fmt.Sprintf(dsnTemplate, s.User, password, s.Host, s.Port,
		s.DBName, sslMode)
}

// PostgresStore is a database store implementation that uses a Postgres
// backend.
type PostgresStore struct {
	cfg *PostgresConfig

	*BaseDB
}

func NewPostgresStore(cfg *PostgresConfig) (*PostgresStore, error) {
	log.Println("Using SQL database '%s'", cfg.DSN(true))

	rawDb, err := sql.Open("postgres", cfg.DSN(false))
	if err != nil {
		return nil, err
	}

	if !cfg.SkipMigrations {
		driver, err := postgres_migrate.WithInstance(
			rawDb, &postgres_migrate.Config{},
		)
		if err != nil {
			return nil, err
		}

		postgresFS := newReplacerFS(sqlSchemas, map[string]string{
			"BLOB":                "BYTEA",
			"INTEGER PRIMARY KEY": "SERIAL PRIMARY KEY",
			"TIMESTAMP":           "TIMESTAMP WITHOUT TIME ZONE",
		})

		err = applyMigrations(
			postgresFS, driver, "migration", cfg.DBName,
		)
		if err != nil {
			return nil, err
		}
	}

	queries := sqlc.New(rawDb)

	return &PostgresStore{
		cfg: cfg,
		BaseDB: &BaseDB{
			DB:      rawDb,
			Queries: queries,
		},
	}, nil
}
