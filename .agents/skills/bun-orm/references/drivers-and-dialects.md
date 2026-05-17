# Drivers And Dialects

Source: https://bun.uptrace.dev/guide/drivers.html

Use this chapter when choosing the database driver, the Bun dialect, or the connection string shape.

## Overview

The driver chapter explains the connection layer Bun sits on top of. Bun does not invent its own transport. It uses a normal Go SQL driver and then layers a dialect over it to account for database-specific SQL behavior.

The practical consequence is that the driver and dialect are a pair. The driver opens the connection and speaks the protocol, while the dialect tells Bun how to render SQL for the target backend.

## Responsibilities

- Driver: opens the connection and handles the protocol.
- Dialect: renders database-specific SQL and scan behavior.
- Bun: builds the query, scans the result, and manages hooks.
- The native `database/sql` handle owns pooling and connection lifetime settings.

## Supported Database Families

- PostgreSQL
- MySQL
- SQLite
- SQL Server
- Oracle

The docs position PostgreSQL as Bun's most complete target, but the other supported databases are first-class enough for practical use.

## Available Features

The official driver page calls out backend-specific feature support. That is the clearest sign that you should not assume all backends behave the same way.

| Feature | Supported Databases |
| --- | --- |
| `InsertOnConflict` | PostgreSQL, SQLite |
| `InsertOnDuplicateKey` | MySQL, MariaDB |
| `InsertReturning` | PostgreSQL, SQLite |
| `UpdateReturning` | PostgreSQL |
| `DeleteReturning` | PostgreSQL |
| `CTE` | PostgreSQL, SQLite, SQL Server |
| `Window` | PostgreSQL, MySQL 8+, SQL Server |

When you write database-specific code, pick the SQL shape that the target database supports instead of forcing one SQL form everywhere.

## Setup Shape

The normal boot flow is:

- open the database with the correct Go driver,
- wrap the `*sql.DB` in `bun.NewDB`,
- choose the matching dialect,
- tune the pool on the native `*sql.DB`,
- keep all of that in bootstrap code rather than in repositories.

## Recommended Bootstrap Pattern

Create the database handles once at application startup and close them on shutdown.

```go
sqldb, err := sql.Open("mysql", dsn)
if err != nil {
	return err
}

db := bun.NewDB(sqldb, mysqldialect.New())
defer db.Close()
```

The production guide also recommends one `sql.DB` and one `bun.DB` per app instance.

## Connection Pool Configuration

The production guide says Bun uses `sql.DB` underneath, so normal Go pool tuning still applies. You should tune the native pool deliberately instead of leaving it at defaults.

Common settings include:

- `SetMaxOpenConns`
- `SetMaxIdleConns`
- `SetConnMaxLifetime`
- `SetConnMaxIdleTime`

The guide also shows a common rule of thumb of using `4 * GOMAXPROCS(0)` as a starting point for `MaxOpenConns` and `MaxIdleConns`, then adjusting for the actual database and workload.

### Example Pool Tuning

```go
func configureDB(db *sql.DB) {
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)
}
```

### Database-Specific Pool Notes

- PostgreSQL: usually supports moderate to high concurrency; watch connection limits.
- MySQL: tune against `max_connections`.
- SQL Server: use a reasonable but not excessive pool.
- SQLite: keep the pool small because SQLite has single-writer limitations.
- Oracle: use a bounded pool and keep connection behavior explicit.

The exact numbers are workload-dependent, but the guide's message is clear: do not rely on defaults in production.

## Database-Specific Behavior

The guide treats backend differences as normal rather than exceptional. When code depends on a feature that is not universal, the recommended approach is to check capability or dialect rather than pretending all databases behave identically.

Examples of backend-sensitive behavior include:

- conflict handling syntax,
- `RETURNING` support,
- CTE support,
- window functions,
- and database-specific expression syntax.

## Direct Dialect Checking

For more complex database-specific logic, check the dialect directly. That lets you keep the query shape honest while still choosing backend-specific hints or expressions.

The pattern is:

- inspect the active dialect,
- branch only when the SQL truly depends on backend behavior,
- and keep backend-specific logic out of generic repository code unless it is unavoidable.

The exact hinting strategy depends on the backend. The important point is to branch on the dialect only when the SQL really depends on backend behavior.

## Common Troubleshooting

The official driver page also mentions production issues to watch for:

- `driver: bad connection` usually means the pool lifetime or connection validation needs attention,
- SQLite lock errors usually mean the concurrency model or pool size needs to be adjusted,
- production startup should validate the connection early rather than discovering it through request traffic.

## Multiple Databases

If the application needs more than one database family, create one Bun instance per database. Keep the dialect and driver matched per instance instead of trying to force a single configuration to fit every backend.

This also applies when different databases are used for different concerns, such as:

- primary OLTP storage,
- read replicas,
- analytics stores,
- or a separate legacy database.

## Production Guidance

- Create the `sql.DB` and `bun.DB` once at startup.
- Close both cleanly on shutdown.
- Configure the pool explicitly.
- Ping the connection early so startup fails fast.
- Use `WithDiscardUnknownColumns` if you need more resilience during rolling migrations.

```go
db := bun.NewDB(sqldb, pgdialect.New(), bun.WithDiscardUnknownColumns())
```

`WithDiscardUnknownColumns` can make production deployments less fragile when schema changes roll out before the application updates fully.

## Representative Example

```go
sqldb, err := sql.Open("mysql", dsn)
if err != nil {
	return err
}

configureDB(sqldb)

db := bun.NewDB(sqldb, mysqldialect.New())
defer db.Close()
```

## Practical Use

- Keep the driver and dialect aligned.
- Do not import a dialect that does not match the driver.
- Place this logic in app bootstrap, not inside repositories.
- Tune the `database/sql` pool deliberately.
- Prefer explicit backend handling when the query depends on non-portable SQL.
- Validate the connection at startup instead of waiting for the first request.
