# Golang Common Table Expressions

Source: https://bun.uptrace.dev/guide/query-common-table-expressions.html

Use this chapter when a query needs named subqueries, reusable derived data, staged updates, or multi-step composition that would otherwise be hard to read.

## Overview

CTEs are Bun's way to keep larger SQL statements readable. The guide shows them as a way to split a complex statement into named logical steps instead of burying everything inside one dense expression.

The main reason to use a CTE in Bun is clarity:

- name the intermediate result,
- reuse it in later parts of the query,
- keep the SQL stages visible,
- and make it easier to explain the query during review.

## WITH

Most Bun queries support CTEs through the `With` method.

```go
query := db.NewSelect().
	With("active_users", db.NewSelect().
		Model((*User)(nil)).
		Column("id").
		Where("status = ?", "active"),
	).
	Table("active_users")
```

You can attach one or more CTEs to select, insert, update, or delete queries. That means CTEs are not a separate query family; they are a composition tool that can be used across the main Bun builders.

## Why CTEs Matter

CTEs are helpful when a query would otherwise become a nested wall of SQL. The official Bun guide uses them for:

- naming intermediate result sets,
- making layered reporting queries easier to follow,
- reusing a subquery in more than one place,
- and making bulk write flows more understandable.

In practice, a CTE is worth it when the query reads better as stages than as one giant expression.

## Subqueries Versus CTEs

Subqueries and CTEs are both valid composition tools, but they solve slightly different readability problems.

- Use a subquery when the derived result only needs to appear in one place.
- Use a CTE when the derived result deserves a name or will be reused.
- Use a CTE when staged SQL is easier to reason about than nested SQL.
- Skip a CTE when the query is clearer without it.

## VALUES

Bun also provides `ValuesQuery` to help build CTEs from in-memory data.

That is useful when:

- you already have the important values in application memory,
- you want to join or update against a small structured dataset,
- and staging the data in a temporary table would be unnecessary overhead.

### Values CTE Example

```go
rows := []struct {
	ID    int64
	Title string
	Text  string
}{
	{ID: 1, Title: "title1", Text: "text1"},
	{ID: 2, Title: "title2", Text: "text2"},
}

err := db.NewUpdate().
	With("_data", db.NewValues(&rows)).
	Model((*Book)(nil)).
	Table("_data").
	Set("title = _data.title").
	Set("text = _data.text").
	Where("book.id = _data.id").
	Exec(ctx)
```

The important idea is that Bun can treat application data as a named relational source inside the SQL statement.

## WithOrder

`WithOrder` lets a values CTE preserve row order by including an order column.

```go
data := []struct {
	ID    int64
	Email string
}{
	{ID: 42, Email: "one@my.com"},
	{ID: 43, Email: "two@my.com"},
}

err := db.NewSelect().
	With("data", db.NewValues(&data).WithOrder()).
	Model((*User)(nil)).
	Where("user.id = data.id").
	OrderExpr("data._order").
	Scan(ctx)
```

This is useful when:

- the input order matters,
- you want the output to reflect the application-side ordering,
- or the query needs deterministic ordering over a small in-memory set.

## Bulk Delete With CTEs

The official guide shows CTEs as a practical way to chunk bulk deletions. That pattern is useful when you want to delete rows that match a predicate but do it in bounded batches.

```go
const batchSize = 1000

for {
	res, err := db.NewDelete().
		With("todo", db.NewSelect().
			Model((*Comment)(nil)).
			Where("created_at < now() - interval '90 day'").
			Limit(batchSize),
		).
		Model((*Comment)(nil)).
		Table("todo").
		Where("comment.id = todo.id").
		Exec(ctx)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected < batchSize {
		break
	}
}
```

The point is not only deletion. The same staging pattern is useful when a large operation should be broken into manageable SQL steps.

## Copying Data Between Tables

The guide also shows using a CTE to copy rows between tables.

```go
err := db.NewInsert().
	With("src", db.NewSelect().
		Model((*Comment)(nil)),
	).
	Table("comments_backup", "src").
	Exec(ctx)
```

This pattern is useful when:

- you want to stage a source query once,
- the destination should consume that source in the same statement,
- or the copy logic is easier to audit as one SQL unit.

## Query Design Notes

- Give each CTE a name that matches the role it plays.
- Keep each CTE small enough to understand in isolation.
- Prefer CTEs when the query reads more clearly as staged steps.
- Skip them when the same SQL is easier to understand in one block.
- Use values CTEs when the important data already exists in memory.
- Use `WithOrder` when the ordering of the value rows matters.

## Representative Example

```go
activeUsers := db.NewSelect().
	Model((*User)(nil)).
	Column("id").
	Where("status = ?", "active")

var users []User
err := db.NewSelect().
	With("active_users", activeUsers).
	Table("active_users").
	Join("JOIN users AS u ON u.id = active_users.id").
	Scan(ctx, &users)
```

```go
rows := []struct {
	ID    int64
	Title string
}{
	{ID: 1, Title: "alpha"},
	{ID: 2, Title: "beta"},
}

err := db.NewSelect().
	With("data", db.NewValues(&rows).WithOrder()).
	Table("data").
	OrderExpr("data._order").
	Scan(ctx)
```

## Practical Use

- Use CTEs when they make the query easier to understand.
- Keep each logical step named and visible.
- Avoid CTEs when a simpler query is clearer.
- Reach for values CTEs when you need to join against application memory.
- Use `WithOrder` when row ordering from the input set matters.
