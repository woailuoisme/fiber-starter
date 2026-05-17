# Database SQL Transactions

Source: https://bun.uptrace.dev/guide/transactions.html

Use this chapter when multi-step writes need to succeed or fail together.

## Overview

Bun's transaction chapter treats transactions as a first-class part of the DB API. `bun.Tx` behaves like a thin wrapper over `sql.Tx`, but it keeps the Bun query helpers available inside the transaction scope.

This chapter is mainly about three things:

- starting transactions correctly,
- running Bun queries inside a transaction,
- and writing reusable persistence helpers that work with both a transaction and a plain database handle.

## Starting Transactions

`bun.Tx` is built on top of `sql.Tx`. You begin a transaction with the database handle, then commit or roll it back just like you would with the standard library.

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
	return err
}
defer tx.Rollback()

if err := tx.Commit(); err != nil {
	return err
}
```

The important behavior is the same as normal SQL transactions:

- the work is isolated within the transaction scope,
- commit makes the changes durable,
- rollback cancels the changes if something fails.

## Running Queries In A Transaction

Once you have a `bun.Tx`, you can use the normal Bun query builder on it.

```go
_, err := tx.NewInsert().
	Model(&user).
	Exec(ctx)
```

```go
_, err := tx.NewUpdate().
	Model(&user).
	Column("name", "updated_at").
	WherePK().
	Exec(ctx)
```

```go
err := tx.NewSelect().
	Model(&user).
	WherePK().
	Scan(ctx)
```

The transaction handle keeps the same query surface as `bun.DB`, so the code does not need a second persistence API just because it is inside a transaction.

## RunInTx

`RunInTx` is the preferred pattern when a unit of work should commit on success and roll back on error.

```go
err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
	if _, err := tx.NewInsert().Model(&user).Exec(ctx); err != nil {
		return err
	}
	if _, err := tx.NewInsert().Model(&profile).Exec(ctx); err != nil {
		return err
	}
	return nil
})
```

The contract is simple:

- return `nil` and the transaction is committed,
- return an error and the transaction is rolled back.

That makes `RunInTx` a good fit for service-layer workflows where one logical business action touches multiple tables.

## IDB Interface

Bun provides `bun.IDB` so the same helper functions can work with `*bun.DB`, `bun.Tx`, and `bun.Conn`.

```go
func InsertUser(ctx context.Context, db bun.IDB, user *User) error {
	_, err := db.NewInsert().Model(user).Exec(ctx)
	return err
}
```

This pattern matters because it lets you write reusable repository helpers:

- direct writes can use the base database handle,
- multi-step workflows can pass a transaction,
- and the same function signature can support both modes.

## Transaction Boundary Discipline

The guide's transaction story is really about boundary placement. Start the transaction at the use-case boundary, perform the database work, and keep outside effects out of the transaction body whenever possible.

Practical rules:

- keep the transaction small,
- keep it around one business unit of work,
- avoid network calls inside the transaction,
- and keep the transaction body deterministic enough to reason about rollback behavior.

## Advisory Locks

The official guide also shows PostgreSQL advisory locks inside a transaction. Use them only when the workflow really needs serialization or coordination that ordinary row locks do not provide.

```go
err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(1)"); err != nil {
		return err
	}
	if _, err := tx.NewSelect().ColumnExpr("pg_advisory_xact_lock(2)").Exec(ctx); err != nil {
		return err
	}
	return nil
})
```

Advisory locks are powerful, but they make the workflow more database-specific. Use them when the coordination problem justifies that tradeoff.

## Hooks And Transactions

`bun.Tx` still supports the Bun hook system. That means transactional work can participate in logging or tracing the same way ordinary queries do. The important point is that hooks should not be used as a way to hide core business rules.

## Representative Example

```go
err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
	if err := repo.CreateUser(ctx, tx, user); err != nil {
		return err
	}
	return repo.CreateProfile(ctx, tx, profile)
})
```

```go
func InsertUser(ctx context.Context, db bun.IDB, user *User) error {
	_, err := db.NewInsert().Model(user).Exec(ctx)
	return err
}
```

```go
err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
	if err := InsertUser(ctx, tx, user1); err != nil {
		return err
	}
	if err := InsertUser(ctx, tx, user2); err != nil {
		return err
	}
	return nil
})
```

## Practical Use

- Put the transaction boundary where the business unit of work begins.
- Keep external side effects outside the transaction whenever possible.
- Use `bun.IDB` for reusable persistence helpers.
- Keep repository helpers transaction-agnostic so they compose naturally.
- Use advisory locks sparingly and only when the workflow requires them.
