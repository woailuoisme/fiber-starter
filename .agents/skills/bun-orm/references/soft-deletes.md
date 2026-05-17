# Soft Deletes

Source: https://bun.uptrace.dev/guide/soft-deletes.html

Use this chapter when the model should keep rows instead of physically deleting them.

## Overview

Soft deletes let Bun mark a row as deleted without actually removing it from the table. The guide shows this as a first-class pattern rather than an ad hoc convention. The core idea is simple: the delete operation becomes an update that records deletion time instead of removing the row.

## How It Works

- Add a timestamp field with the `soft_delete` tag.
- Ordinary deletes become updates to that timestamp field.
- Normal selects exclude deleted rows.
- Deleted rows can be queried explicitly when needed.
- Force delete is available when a real delete is required.
- The deleted timestamp is part of the model contract, not just an implementation detail.
- Use `nullzero` when you want the zero time to behave like `NULL`.
- The deleted timestamp usually serves as the flag for both state and deletion time.

The official examples show this pattern with a timestamp column such as `deleted_at`. Once that column is tagged, Bun treats the row as live unless the timestamp is populated.

## Marking A Row As Deleted

The soft-delete flow works by turning the delete into an update on the marker column.

```go
_, err := db.NewDelete().
	Model((*User)(nil)).
	Where("id = ?", 123).
	Exec(ctx)
```

Under soft delete, that operation updates the soft-delete timestamp rather than physically removing the row.

If you want the equivalent SQL shape in your head, think of it as:

- set the delete marker,
- keep the row in the table,
- and let future normal selects ignore it.

## Query Variants

- `WhereDeleted` returns only soft-deleted rows.
- `WhereAllWithDeleted` includes both live and deleted rows.
- `ForceDelete` bypasses the soft-delete behavior and removes the row physically.
- Normal `SELECT` queries automatically ignore soft-deleted rows unless you ask for them explicitly.

These variants are the main operational API for soft delete:

- use `WhereDeleted` when you only want archived rows,
- use `WhereAllWithDeleted` when you need the full history view,
- use `ForceDelete` when the row must be physically removed.

## Select Behavior

Normal selects hide soft-deleted rows by default. That is the main behavioral difference from a normal timestamp column.

```go
err := db.NewSelect().
	Model(&users).
	Scan(ctx)
```

If you want only deleted rows:

```go
err := db.NewSelect().
	Model(&users).
	WhereDeleted().
	Scan(ctx)
```

If you want both live and deleted rows:

```go
err := db.NewSelect().
	Model(&users).
	WhereAllWithDeleted().
	Scan(ctx)
```

That query behavior is why soft delete is not just a model tag. It changes the default visibility rules for reads as well as writes.

## Index Planning

The unique-index warning in the guide is important: a soft-deleted row still exists, so a unique constraint still sees it unless the index definition accounts for deletion state. Plan the index before you rely on soft delete in production.

## Unique Index Caution

The docs point out a common issue: soft-deleted rows still exist, so uniqueness rules can still conflict unless you design the index strategy carefully. That means the database schema and the soft-delete policy have to be planned together.

The practical options are:

- partial unique indexes that ignore deleted rows where the database supports them,
- composite strategies that include the deleted timestamp,
- or a different business key if the deleted rows must stay visible to uniqueness checks.

The official guide also notes that if your database cannot index expressions the same way, you may need to avoid `nullzero` and let Bun write a zero timestamp explicitly so the unique-key strategy stays workable.

## Force Delete

`ForceDelete` is the escape hatch when you truly want to remove the row physically. It bypasses the soft-delete behavior and emits a real delete.

Use it sparingly:

- for cleanup jobs,
- for compliance-sensitive permanent deletion,
- or for cases where the row should no longer exist at all.

Do not use force delete as a default just because soft delete exists. The point of soft delete is to make the normal path preserve the record.

## Representative Example

```go
type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`
	DeletedAt     time.Time `bun:",soft_delete,nullzero"`
}
```

```go
_, err := db.NewDelete().
	Model((*User)(nil)).
	Where("id = ?", 123).
	Exec(ctx)
```

```go
err := db.NewSelect().
	Model(&users).
	WhereDeleted().
	Scan(ctx)
```

```go
err := db.NewSelect().
	Model(&users).
	WhereAllWithDeleted().
	Scan(ctx)
```

```go
_, err := db.NewDelete().
	Model((*User)(nil)).
	Where("id = ?", 123).
	ForceDelete().
	Exec(ctx)
```

## Practical Use

- Use soft deletes when recovery or auditability matters.
- Decide how unique indexes should behave before shipping.
- Use force delete only when the application really wants to remove the row.
- Make the delete policy visible in the model and the query code.
- Treat soft delete as part of the schema contract, not a hidden runtime trick.
