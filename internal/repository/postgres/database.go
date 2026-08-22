package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wyw14/cry-081/internal/domain/shared"
)

type queryer interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type transactionKey struct{}

type Database struct{ pool *pgxpool.Pool }

func Open(ctx context.Context, databaseURL string) (*Database, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database configuration: %w", err)
	}
	config.MaxConns = 10
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &Database{pool: pool}, nil
}

func (d *Database) Close() { d.pool.Close() }

func (d *Database) Ping(ctx context.Context) error { return d.pool.Ping(ctx) }

func (d *Database) Within(ctx context.Context, callback func(context.Context) error) error {
	tx, err := d.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return translate(err)
	}
	txContext := context.WithValue(ctx, transactionKey{}, tx)
	if err := callback(txContext); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return translate(tx.Commit(ctx))
}

func (d *Database) queries(ctx context.Context) queryer {
	if tx, ok := ctx.Value(transactionKey{}).(pgx.Tx); ok {
		return tx
	}
	return d.pool
}

func translate(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return shared.ErrNotFound
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23505", "23503", "23514":
			return shared.NewError("DATABASE_CONSTRAINT", "database constraint rejected the change", shared.ErrConflict)
		case "40001":
			return shared.ErrVersionConflict
		}
	}
	return err
}
