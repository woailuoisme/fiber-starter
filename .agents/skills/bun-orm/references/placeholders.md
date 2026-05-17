# SQL Placeholders

Source: https://bun.uptrace.dev/guide/placeholders.html

Use this chapter when composing SQL with dynamic values, identifiers, lists, or reusable model metadata.

## Overview

Bun's placeholder system is the safety layer that keeps SQL structure separate from user-controlled data. The general rule is simple:

- values are bound as values,
- identifiers are quoted as identifiers,
- trusted fragments are only used when the application owns the SQL text,
- and model metadata can be reused as placeholders when the query needs to stay in sync with the model.

This makes Bun practical for dynamic queries without encouraging unsafe string concatenation.

## Basic And Positional Placeholders

The normal `?` placeholder binds values safely.

```go
err := db.NewSelect().
	TableExpr("users").
	ColumnExpr("?, ?", "foo", "bar").
	Scan(ctx)
```

You can also use positional placeholders when the same value needs to appear more than once.

```go
err := db.NewSelect().
	TableExpr("users").
	ColumnExpr("?0, ?1, ?0", "foo", "bar").
	Scan(ctx)
```

Use these rules consistently:

- bind all runtime values with placeholders,
- let Bun quote and escape string values,
- keep SQL fragments readable,
- and do not build SQL by concatenating request data.

## bun.Ident

Use `bun.Ident` when you need a quoted identifier such as a table name, column name, schema name, or alias that must appear in SQL syntax rather than as a value.

```go
err := db.NewSelect().
	TableExpr("users").
	ColumnExpr("? = ?", bun.Ident("status"), "active").
	Scan(ctx)
```

`bun.Ident` is the correct choice when:

- the database expects an identifier, not a string literal,
- the identifier comes from a narrow allow-list,
- or you want Bun to quote the identifier correctly for the current dialect.

Do not use it as a shortcut for arbitrary user input. The identifier still needs to be controlled by application logic.

## bun.Safe

`bun.Safe` disables quoting entirely. That means the fragment is inserted as trusted SQL text.

```go
err := db.NewSelect().
	TableExpr("(?) AS gs", bun.Safe("generate_series(0, 10)")).
	Scan(ctx)
```

This is useful when:

- the SQL fragment is fully owned by the application,
- the expression is not something Bun should escape as a value,
- or the database expression must stay exactly as written.

Because `bun.Safe` bypasses Bun's normal escaping model, treat it as an exception. If the text is not fully trusted, do not use it.

## bun.In

`bun.In` expands slices into `IN (...)` lists and keeps the values bound safely.

```go
err := db.NewSelect().
	TableExpr("users").
	Where("id IN (?)", bun.In([]int64{1, 2, 3})).
	Scan(ctx)
```

Composite keys can use nested slices.

```go
err := db.NewSelect().
	TableExpr("pairs").
	Where("(foo, bar) IN (?)", bun.In([][]string{
		{"hello", "world"},
		{"hell", "yeah"},
	})).
	Scan(ctx)
```

Use `bun.In` when:

- you need to filter by a dynamic list,
- the list is coming from application data rather than SQL text,
- or you want to avoid manually expanding placeholders.

## Model Placeholders

Bun also exposes placeholders that expand to metadata from the current model.

- `?TableName` gives the model table name.
- `?TableAlias` gives the current table alias.
- `?PKs` gives the model's primary key columns.
- `?TablePKs` gives the primary key columns with the table alias.
- `?Columns` gives the model columns.
- `?TableColumns` gives the model columns with the table alias.

These placeholders are helpful when:

- you want a query to stay aligned with the model definition,
- you want to avoid repeating column lists,
- or the query needs to reference model metadata in a reusable way.

```go
err := db.NewSelect().
	Model((*User)(nil)).
	ColumnExpr("?TableColumns").
	Scan(ctx)
```

In practice, model placeholders are useful for:

- generic helpers,
- reusable query fragments,
- model-driven projection,
- and SQL that should stay correct when the model's schema changes.

## Global Placeholders

Bun also supports named global placeholders. These are useful when you want one query template to adapt to different schema names or other trusted structural values.

```go
db1 := db.WithNamedArg("SCHEMA", bun.Ident("foo"))
db2 := db.WithNamedArg("SCHEMA", bun.Ident("bar"))

err := db1.NewSelect().TableExpr("?SCHEMA.table").Scan(ctx)
err = db2.NewSelect().TableExpr("?SCHEMA.table").Scan(ctx)
```

This is a good fit for:

- multi-tenant schema prefixes,
- different query contexts with shared SQL templates,
- and application-owned SQL fragments that should remain reusable.

## Safety Rules

- Bind runtime values with placeholders.
- Use `bun.Ident` only for controlled identifiers.
- Use `bun.Safe` only for trusted SQL text.
- Keep `IN` lists in `bun.In`.
- Keep model placeholders aligned with the current model definition.
- Prefer the simplest placeholder that expresses the intent clearly.

## Common Uses

- dynamic table or column selection from a closed set,
- `IN` filters from a list of IDs or keys,
- reusable query fragments,
- schema or tenant names supplied by trusted application configuration,
- and model-aware SQL that should stay aligned with Bun metadata.

## Representative Example

```go
err := db.NewSelect().
	Model((*User)(nil)).
	Where("status = ?", status).
	Where("id IN (?)", bun.In([]int64{1, 2, 3})).
	OrderExpr("? DESC", bun.Ident("created_at")).
	Scan(ctx)
```

```go
err := db.NewSelect().
	TableExpr("?SCHEMA.users").
	Scan(ctx)
```

## Practical Use

- Prefer bound values over SQL fragments.
- Keep identifier allow-lists narrow.
- Review `bun.Safe` usage carefully because it bypasses the normal escaping model.
- Use model placeholders when they reduce drift between the query and the model.
- Use global placeholders when a trusted application setting should reshape the SQL template.
