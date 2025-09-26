package store

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pickeringtech/FinOpsAggregator/internal/config"
	"github.com/rs/zerolog/log"
)

// DB wraps the database connection and provides query building
type DB struct {
	pool *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

// NewDB creates a new database connection
func NewDB(cfg config.PostgresConfig) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().Msg("Database connection established")

	return &DB{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}, nil
}

// Close closes the database connection
func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
		log.Info().Msg("Database connection closed")
	}
}

// Pool returns the underlying connection pool
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// QueryBuilder returns a new query builder
func (db *DB) QueryBuilder() squirrel.StatementBuilderType {
	return db.sb
}

// WithTx executes a function within a database transaction
func (db *DB) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				log.Error().Err(rollbackErr).Msg("Failed to rollback transaction after panic")
			}
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			log.Error().Err(rollbackErr).Msg("Failed to rollback transaction")
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Store provides access to all repositories
type Store struct {
	db    *DB
	Nodes *NodeRepository
	Edges *EdgeRepository
	Costs *CostRepository
	Usage *UsageRepository
	Runs  *RunRepository
}

// NewStore creates a new store with all repositories
func NewStore(db *DB) *Store {
	return &Store{
		db:    db,
		Nodes: NewNodeRepository(db),
		Edges: NewEdgeRepository(db),
		Costs: NewCostRepository(db),
		Usage: NewUsageRepository(db),
		Runs:  NewRunRepository(db),
	}
}

// DB returns the underlying database connection
func (s *Store) DB() *DB {
	return s.db
}

// WithTx executes a function within a database transaction
func (s *Store) WithTx(ctx context.Context, fn func(*Store) error) error {
	return s.db.WithTx(ctx, func(tx pgx.Tx) error {
		txStore := &Store{
			db:    &DB{pool: nil, sb: s.db.sb}, // We'll use tx directly
			Nodes: NewNodeRepositoryWithTx(tx, s.db.sb),
			Edges: NewEdgeRepositoryWithTx(tx, s.db.sb),
			Costs: NewCostRepositoryWithTx(tx, s.db.sb),
			Usage: NewUsageRepositoryWithTx(tx, s.db.sb),
			Runs:  NewRunRepositoryWithTx(tx, s.db.sb),
		}
		return fn(txStore)
	})
}

// Queryable interface for both pool and transaction
type Queryable interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

// BaseRepository provides common functionality for all repositories
type BaseRepository struct {
	db Queryable
	sb squirrel.StatementBuilderType
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db Queryable, sb squirrel.StatementBuilderType) *BaseRepository {
	return &BaseRepository{
		db: db,
		sb: sb,
	}
}

// QueryBuilder returns the statement builder
func (r *BaseRepository) QueryBuilder() squirrel.StatementBuilderType {
	return r.sb
}

// DB returns the queryable database interface
func (r *BaseRepository) DB() Queryable {
	return r.db
}

// ExecQuery executes a query built with squirrel
func (r *BaseRepository) ExecQuery(ctx context.Context, query squirrel.Sqlizer) (pgconn.CommandTag, error) {
	sql, args, err := query.ToSql()
	if err != nil {
		return pgconn.CommandTag{}, fmt.Errorf("failed to build query: %w", err)
	}

	log.Debug().
		Str("sql", sql).
		Interface("args", args).
		Msg("Executing query")

	return r.db.Exec(ctx, sql, args...)
}

// QueryRows executes a query and returns rows
func (r *BaseRepository) QueryRows(ctx context.Context, query squirrel.Sqlizer) (pgx.Rows, error) {
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	log.Debug().
		Str("sql", sql).
		Interface("args", args).
		Msg("Executing query")

	return r.db.Query(ctx, sql, args...)
}

// QueryRow executes a query and returns a single row
func (r *BaseRepository) QueryRow(ctx context.Context, query squirrel.Sqlizer) pgx.Row {
	sql, args, err := query.ToSql()
	if err != nil {
		log.Error().Err(err).Msg("Failed to build query")
		// Return a row that will error when scanned
		return r.db.QueryRow(ctx, "SELECT 1 WHERE FALSE")
	}

	log.Debug().
		Str("sql", sql).
		Interface("args", args).
		Msg("Executing query")

	return r.db.QueryRow(ctx, sql, args...)
}
