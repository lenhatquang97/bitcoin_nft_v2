package db

import (
	"bitcoin_nft_v2/db/sqlc"
	"context"
	"database/sql"
)

type TxOptions interface {
	ReadOnly() bool
}

type BatchedTx[Q any] interface {
	ExecTx(ctx context.Context, txOptions TxOptions,
		txBody func(Q) error) error
}

type Tx interface {
	Commit() error

	Rollback() error
}

type QueryCreator[Q any] func(*sql.Tx) Q

type BatchedQuerier interface {
	sqlc.Querier

	BeginTx(ctx context.Context, options TxOptions) (*sql.Tx, error)
}

type TransactionExecutor[Query any] struct {
	BatchedQuerier

	createQuery QueryCreator[Query]
}

func NewTransactionExecutor[Querier any](db BatchedQuerier,
	createQuery QueryCreator[Querier]) *TransactionExecutor[Querier] {

	return &TransactionExecutor[Querier]{
		BatchedQuerier: db,
		createQuery:    createQuery,
	}
}

func (t *TransactionExecutor[Q]) ExecTx(ctx context.Context,
	txOptions TxOptions, txBody func(Q) error) error {

	// Create the db transaction.
	tx, err := t.BatchedQuerier.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}

	// Rollback is safe to call even if the tx is already closed, so if the
	// tx commits successfully, this is a no-op.
	defer func() {
		_ = tx.Rollback()
	}()

	if err := txBody(t.createQuery(tx)); err != nil {
		return err
	}

	// Commit transaction.
	//
	// TODO(roasbeef): need to handle SQLITE_BUSY here?
	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

type BaseDB struct {
	*sql.DB

	*sqlc.Queries
}

func (s *BaseDB) BeginTx(ctx context.Context, opts TxOptions) (*sql.Tx, error) {
	sqlOptions := sql.TxOptions{
		ReadOnly: opts.ReadOnly(),
	}
	return s.DB.BeginTx(ctx, &sqlOptions)
}
