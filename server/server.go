package server

import (
	"bitcoin_nft_v2/db"
)

type Server struct {
	PostgresDB *db.PostgresStore
}

func InitServer() (*Server, error) {
	sqlFixture := db.NewTestPgFixture()
	store, err := db.NewPostgresStore(sqlFixture.GetConfig())
	if err != nil {
		return nil, err
	}

	return &Server{
		PostgresDB: store,
	}, nil
}
