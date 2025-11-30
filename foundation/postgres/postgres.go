package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type Db struct {
	db     *sqlx.DB
	config Config
}

func NewPostgres(config Config) (*Db, error) {
	pg := &Db{config: config}
	if err := pg.connect(); err != nil {
		return nil, err
	}

	return pg, nil
}

func (p *Db) connect() error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		p.config.Host, p.config.Port, p.config.User, p.config.Password, p.config.DBName)
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return err
	}
	p.db = db
	return nil
}

func (p *Db) GetDB() *sqlx.DB {
	return p.db
}

func (p *Db) Init(folder string) error {
	db := p.GetDB().DB

	err := goose.SetDialect("postgres")
	if err != nil {
		return err
	}

	// Goose up
	return goose.Up(db, folder)
}

func (p *Db) Close() error {
	return p.db.Close()
}

// QueryList is a generic function that retrieves a list of T from the database using the provided query and args.
func QueryList[T any](ctx context.Context, db *sqlx.DB, query string, args ...any) ([]T, error) {
	var results []T
	err := db.SelectContext(ctx, &results, query, args...)
	return results, err
}

// QueryOne is a generic function that retrieves a single T from the database using the provided query and args.
func QueryOne[T any](ctx context.Context, db *sqlx.DB, query string, args ...any) (*T, error) {
	var result T
	err := db.GetContext(ctx, &result, query, args...)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ExecQuery is a generic function that executes a query with the provided args.
func ExecQuery(ctx context.Context, db *sqlx.DB, query string, args ...any) (sql.Result, error) {
	return db.ExecContext(ctx, query, args...)
}
