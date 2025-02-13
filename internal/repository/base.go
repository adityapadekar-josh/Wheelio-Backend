package repository

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
)

type BaseRepository struct {
	DB *sql.DB
}

type RepositoryTransaction interface {
	BeginTx(ctx context.Context) (*sql.Tx, error)
	CommitTx(tx *sql.Tx) error
	RollbackTx(tx *sql.Tx) error
	HandleTransaction(ctx context.Context, tx *sql.Tx, incomingErr error) error
}

type QueryExecuter interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func (b *BaseRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	tx, err := b.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("error occurred while initiating database transaction", "error", err)
		return nil, apperrors.ErrInternalServer
	}

	return tx, nil
}

func (b *BaseRepository) CommitTx(tx *sql.Tx) error {
	err := tx.Commit()
	if err != nil {
		slog.Error("error occurred while committing database transaction", "error", err)
		return apperrors.ErrInternalServer
	}

	return nil
}

func (b *BaseRepository) RollbackTx(tx *sql.Tx) error {
	err := tx.Rollback()
	if err != nil {
		slog.Error("error occurred while rolling back database transaction", "error", err)
		return apperrors.ErrInternalServer
	}

	return nil
}

func (b *BaseRepository) HandleTransaction(ctx context.Context, tx *sql.Tx, incomingErr error) error {
	if incomingErr != nil {
		err := tx.Rollback()
		if err != nil {
			slog.Error("error occurred while rolling back database transaction", "error", err)
			return apperrors.ErrInternalServer
		}
		return nil
	}

	err := tx.Commit()
	if err != nil {
		slog.Error("error occurred while committing database transaction", "error", err)
		return apperrors.ErrInternalServer
	}
	return nil
}

func (b *BaseRepository) initiateQueryExecuter(tx *sql.Tx) QueryExecuter {
	if tx != nil {
		return tx
	}

	return b.DB
}
