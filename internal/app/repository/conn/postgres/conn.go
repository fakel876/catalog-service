package rcpostgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"

	"github.com/fakel876/catalog-service/internal/app/config/section"
	"github.com/fakel876/catalog-service/migration"
)

type (
	Client struct {
		_bunDB
		rawbunDB *bun.DB

		cfg section.RepositoryPostgres
	}
	_bunDB = bun.IDB
)

func (c *Client) GetRawBunDB() *bun.DB {
	return c.rawbunDB
}

func NewClient(ctx context.Context, cfg section.RepositoryPostgres) (*Client, error) {
	var u url.URL
	u.Scheme = "postgres"
	u.Host = cfg.Address
	u.User = url.UserPassword(cfg.Username, cfg.Password)
	u.Path = cfg.Name

	args := make(url.Values)
	args.Add("sslmode", "disable")
	u.RawQuery = args.Encode()

	dsn := u.String()

	log.Printf("read timeout: %v, write timeout: %v", cfg.ReadTimeout, cfg.WriteTimeout)

	sqlDB := sql.OpenDB(
		pgdriver.NewConnector(
			pgdriver.WithDSN(dsn),
			pgdriver.WithReadTimeout(cfg.ReadTimeout),
			pgdriver.WithWriteTimeout(cfg.WriteTimeout),
		),
	)
	sqlDB.SetMaxOpenConns(10)

	db := bun.NewDB(sqlDB, pgdialect.New(), bun.WithDiscardUnknownColumns())

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Client{db, db, cfg}, nil
}

func (c *Client) Migrate(ctx context.Context) (oldVer, newVer int64, err error) {
	migrations := migrate.NewMigrations()

	if err := migrations.Discover(migration.Postgres); err != nil {
		return 0, 0, fmt.Errorf("failed to discover migrations: %w", err)
	}

	opts := []migrate.MigratorOption{
		migrate.WithTableName(c.cfg.MigrationTable),
		migrate.WithLocksTableName(c.cfg.MigrationTable + "_lock"),
		migrate.WithMarkAppliedOnSuccess(true),
	}

	migrator := migrate.NewMigrator(c.rawbunDB, migrations, opts...)

	if err := migrator.Init(ctx); err != nil {
		return 0, 0, err
	}

	applied, err := migrator.AppliedMigrations(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	for _, mg := range applied {
		v, _ := strconv.ParseInt(mg.Name, 10, 64)
		if v > oldVer {
			oldVer = v
		}
	}

	newVer = oldVer

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return oldVer, newVer, fmt.Errorf("failed to apply migrations: %w", err)
	}

	for _, mg := range group.Migrations {
		v, _ := strconv.ParseInt(mg.Name, 10, 64)
		if v > newVer {
			newVer = v
		}
	}

	return oldVer, newVer, nil
}
