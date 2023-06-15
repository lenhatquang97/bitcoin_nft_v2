package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

const (
	testPgUser   = "root"
	testPgPass   = "secret"
	testPgDBName = "nft_collection"
	PostgresTag  = "11"
)

type TestPgFixture struct {
	db       *sql.DB
	pool     *dockertest.Pool
	resource *dockertest.Resource
	host     string
	port     int
}

func NewTestPgFixture() *TestPgFixture {
	pool, err := dockertest.NewPool("")

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        PostgresTag,
		Env: []string{
			fmt.Sprintf("POSTGRES_USER=%v", testPgUser),
			fmt.Sprintf("POSTGRES_PASSWORD=%v", testPgPass),
			fmt.Sprintf("POSTGRES_DB=%v", testPgDBName),
			"listen_addresses='*'",
		},
		Cmd: []string{
			"postgres",
			"-c", "log_statement=all",
			"-c", "log_destination=stderr",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	fmt.Println(err)
	hostAndPort := resource.GetHostPort("5432/tcp")
	parts := strings.Split(hostAndPort, ":")
	host := parts[0]
	port := 5432

	fixture := &TestPgFixture{
		host: host,
		port: int(port),
	}
	databaseURL := fixture.GetDSN()
	log.Println("Connecting to Postgres fixture: %v\n", databaseURL)

	pool.MaxWait = 120 * time.Second

	var testDB *sql.DB
	err = pool.Retry(func() error {
		testDB, err = sql.Open("postgres", databaseURL)
		if err != nil {
			return err
		}
		return testDB.Ping()
	})

	fixture.db = testDB
	fixture.pool = pool
	fixture.resource = resource

	return fixture
}

func (f *TestPgFixture) GetDSN() string {
	return f.GetConfig().DSN(false)
}

func (f *TestPgFixture) GetConfig() *PostgresConfig {
	return &PostgresConfig{
		Host:       f.host,
		Port:       f.port,
		User:       testPgUser,
		Password:   testPgPass,
		DBName:     testPgDBName,
		RequireSSL: false,
	}
}
