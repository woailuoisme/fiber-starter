# Defining Models

Source: https://bun.uptrace.dev/guide/models.html

Use this chapter when creating or reviewing Bun model structs.

## Overview

Models are the bridge between Go structs and database tables. Bun keeps this mapping explicit so the code remains understandable even when the schema is non-trivial. The guide emphasizes that Bun models should mirror the database, not hide it.

The model chapter is also where Bun's schema naming, aliasing, null handling, and relation tags become visible. If the model is wrong, almost every other query chapter becomes harder to use correctly.

## Core Rules

- Use exported fields for columns.
- Use `bun.BaseModel` to define table name and alias.
- Let Bun map names with snake_case defaults unless the schema needs explicit overrides.
- Keep nullability and defaults explicit in tags or types.
- Keep the model close to the table shape.
- Use DTOs or separate scan structs for non-table responses.
- Keep schema-relevant behavior in tags rather than in hidden helper code.

## Table And Column Names

Bun converts struct names and field names to table and column names by default. That keeps simple models short, but the guide also makes clear that you should override names when the schema already exists or when the default mapping is not the right contract.

Use table-level tags when you need:

- a specific table name,
- a stable alias,
- a different table source for select queries,
- or consistent naming across query branches.

Use field-level tags when you need:

- a custom column name,
- a unique constraint,
- a concrete SQL type,
- or a special behavior such as soft delete or scan-only handling.

## Common Tags

- `table:` sets the table name.
- `alias:` sets the query alias.
- `pk` marks the primary key.
- `autoincrement` marks generated identifiers.
- `notnull` marks required fields.
- `unique` creates a uniqueness constraint.
- `default:` adds a default SQL expression.
- `nullzero` converts Go zero values to SQL `NULL`.
- `scanonly` allows scan-only fields.
- `soft_delete` enables logical deletes.
- `type:` pins a concrete SQL type when the default inference is not enough.
- `rel:` marks relation fields instead of scalar columns.
- `jsonb`, `array`, and other database-specific type helpers keep special columns explicit.

## Nullability Patterns

The guide recommends choosing the representation deliberately:

- pointers when `nil` should mean missing,
- `sql.Null*` or Bun null types when validity must be tracked explicitly,
- `nullzero` when a zero value should be treated as database null/default behavior.

Use pointers when a field is optional in API payloads and should round-trip as missing. Use Bun null helpers when scans and writes need a stronger distinction between missing and zero. Avoid mixing `nullzero` and pointer semantics on the same column without documenting the intent.

## Table And View Targets

`ModelTableExpr` lets a model point at a different table expression while keeping the model shape intact. The official guide treats that as useful for things like:

- partitioned tables,
- views,
- A/B or environment-specific table sets,
- and other cases where the row shape stays the same but the source changes.

The important detail is that aliases still need to match the model's expectations. If the same model is used against different table expressions, the aliases should remain consistent so relation loading and SQL generation keep working.

## Working With NULL Values

The model guide calls out several ways to represent nullable data. The practical rule is to match the Go type to the semantic meaning of the column.

- Use pointer fields when the absence of a value matters.
- Use dedicated null wrappers when scanning and writing should preserve validity explicitly.
- Use `nullzero` when a zero value should behave as database null/default behavior.

This is especially important for text, timestamps, numeric fields, and soft-delete columns.

## Timestamps And Soft Delete

The models page shows common patterns for created/updated timestamps and soft-delete fields. The key point is to decide whether the timestamp is owned by the database default, by hooks, or by application code, and then keep that policy consistent.

If the database owns the timestamp, keep the default expression on the column and let Bun scan it back after insert or update. If the application owns it, set the field before saving and keep the migration default aligned with that convention.

For soft delete, a nullable timestamp is usually more expressive than a boolean because it records both the deleted state and the time of deletion.

## Relationships On Models

The model chapter also shows how relation tags belong on fields that represent related rows, not on plain scalar columns.

- `rel:belongs-to` for a parent reference.
- `rel:has-one` for a single child row.
- `rel:has-many` for a collection.
- many-to-many tags on slice fields when an intermediary table exists.

The important idea is that the model defines the relationship contract, and the query decides when to load it.

## Schema Alignment

Bun model tags are easiest to maintain when they mirror the migration. A model change should usually be reviewed together with the corresponding table change so the following stay consistent:

- primary keys,
- nullability,
- defaults,
- unique constraints,
- naming,
- and soft-delete behavior.

That consistency is important because Bun deliberately keeps the mapping visible rather than hiding schema drift behind a large ORM abstraction.

## Representative Example

```go
type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID        int64       `bun:"id,pk,autoincrement"`
	Email     string      `bun:"email,unique,notnull"`
	Name      string      `bun:"name,notnull"`
	Bio       *string     `bun:"bio"`
	Status    string      `bun:"status,type:varchar(20),default:'active'"`
	CreatedAt time.Time   `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time   `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
	DeletedAt bun.NullTime `bun:",soft_delete,nullzero"`
}
```

```go
type Profile struct {
	bun.BaseModel `bun:"table:profiles,alias:p"`

	ID     int64 `bun:"id,pk,autoincrement"`
	UserID int64 `bun:"user_id,notnull"`
	User   *User `bun:"rel:belongs-to,join:user_id=id"`
}
```

## Advanced Mapping Notes

- Use `alias:` when the same table appears more than once in a query graph.
- Use `scanonly` for derived fields that should appear in reads but not in inserts or updates.
- Use specialized SQL types only when the database type is semantically important.
- Treat array and JSON columns as part of the model contract so query authors know how Bun will encode them.
- Keep model definitions small enough that the schema contract can be reviewed at a glance.

## Practical Use

- Keep model tags aligned with migrations.
- Avoid using models as a place for unrelated business logic.
- Use separate DTOs when the query result shape is not a table shape.
- Review every model change together with the schema migration and the queries that depend on it.
- Keep the schema contract readable enough that a future engineer can inspect it without guessing.
