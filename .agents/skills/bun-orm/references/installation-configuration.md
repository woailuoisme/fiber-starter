# Installation And Configuration

Source:
- https://bun.uptrace.dev/guide/golang-orm.html#installation
- https://bun.uptrace.dev/guide/golang-orm.html#connection-pool-configuration
- https://bun.uptrace.dev/postgres/#supported-drivers

Use this chapter when setting up Bun for a project, choosing the PostgreSQL driver, or tuning the initial connection pool.

## Overview

The official Bun setup path is straightforward:

- install Bun itself,
- install the driver that matches the target database,
- create one `sql.DB` and one `bun.DB`,
- choose the matching dialect,
- and tune the underlying pool deliberately.

This chapter combines the Bun installation page, production connection-pool guidance, and PostgreSQL driver support notes into one setup reference.

## Installation

Install the core Bun package and the driver for the database you plan to use.

```bash
go get github.com/uptrace/bun@latest
go get github.com/uptrace/bun/driver/pgdriver
go get github.com/uptrace/bun/driver/sqliteshim
go get github.com/go-sql-driver/mysql
go get github.com/denisenkom/go-mssqldb
```

The important idea is that Bun itself is only one part of the setup. You also need the database driver and the matching dialect package.

## Database Driver And Dialect Pairing

When you open a database, think in pairs:

- the Go driver handles the connection protocol,
- the Bun dialect renders SQL for that backend.

Examples of the common pairings are:

- PostgreSQL: `pgdriver` + `pgdialect`
- MySQL: `go-sql-driver/mysql` + `mysqldialect`
- SQLite: `sqliteshim` + `sqlitedialect`
- SQL Server: the SQL Server driver + `mssqldialect`

Keep the driver and dialect aligned. If the backend changes, update both together.

## PostgreSQL Supported Drivers

Bun's PostgreSQL guide documents two supported PostgreSQL driver paths.

### pgdriver

`pgdriver` is Bun's own PostgreSQL driver. It connects with a DSN and is the most direct Bun-native option for PostgreSQL.

Use it when you want:

- a Bun-native PostgreSQL driver,
- a direct DSN-based setup,
- and a clean pairing with `pgdialect`.

### pgx

Bun also supports `pgx` through the standard library connection path.

Use it when:

- your project already uses `pgx`,
- you want to keep an existing PostgreSQL driver stack,
- or your team already standardizes on `pgx` for PostgreSQL access.

The practical rule is that both paths are valid. Pick the one that best matches the rest of the repository's database stack.

## Recommended Startup Shape

The official production guidance is to create one `sql.DB` and one `bun.DB` when the app starts, then close them when the app exits.

```go
sqldb, err := sql.Open("mysql", dsn)
if err != nil {
	return err
}

db := bun.NewDB(sqldb, mysqldialect.New())
defer db.Close()
```

The same shape applies to PostgreSQL and SQLite: open the native connection first, wrap it with Bun, and keep the handles at process scope.

## Connection Pool Configuration

Bun relies on `database/sql` for pooling, so the underlying `*sql.DB` must be tuned explicitly.

Common pool settings:

- `SetMaxOpenConns`
- `SetMaxIdleConns`
- `SetConnMaxLifetime`
- `SetConnMaxIdleTime`

### Example Pool Tuning

```go
sqldb.SetMaxOpenConns(25)
sqldb.SetMaxIdleConns(25)
sqldb.SetConnMaxLifetime(5 * time.Minute)
sqldb.SetConnMaxIdleTime(5 * time.Minute)
```

The Bun guide and production page both stress that production code should not rely on defaults. Start with a deliberate configuration and adjust based on the database and traffic pattern.

### Database-Specific Pool Notes

- PostgreSQL: moderate to high concurrency is common, but watch total connection limits.
- MySQL: size the pool with `max_connections` in mind.
- SQL Server: keep the pool reasonable and bounded.
- SQLite: keep concurrency low because SQLite has single-writer constraints.

## Production Safety

The production guide also recommends failing fast at startup by pinging the connection early.

```go
if err := sqldb.Ping(); err != nil {
	return err
}
```

That keeps connection failures out of the request path and makes deployment problems easier to diagnose.

## Rolling Schema Changes

When schema changes roll out gradually, Bun can be configured to tolerate unknown columns in some production scenarios.

```go
db := bun.NewDB(sqldb, pgdialect.New(), bun.WithDiscardUnknownColumns())
```

Use that only when it fits the deployment strategy. It is a resilience measure, not a replacement for proper schema review.

## Practical Setup Checklist

- Install Bun and the driver for the chosen backend.
- Choose the matching dialect.
- Open the native `*sql.DB` first.
- Wrap it with `bun.NewDB`.
- Tune the pool explicitly.
- Ping the database on startup.
- Close the handles on shutdown.
- Keep the driver and dialect together in bootstrap code.

## Representative Example

```go
sqldb, err := sql.Open("mysql", dsn)
if err != nil {
	return err
}

sqldb.SetMaxOpenConns(25)
sqldb.SetMaxIdleConns(25)
sqldb.SetConnMaxLifetime(5 * time.Minute)

db := bun.NewDB(sqldb, mysqldialect.New())
defer db.Close()
```

## Practical Use

- Keep installation, driver selection, and pool tuning in one bootstrap step.
- Use Bun's PostgreSQL driver when you want the simplest Bun-native setup.
- Use `pgx` when the repository already standardizes on it.
- Tune the pool before production traffic reaches the app.
- Treat connection setup as part of runtime configuration, not repository code.
