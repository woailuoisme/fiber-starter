# CRUD And Filters

Source:
- https://bun.uptrace.dev/guide/query-select.html
- https://bun.uptrace.dev/guide/query-insert.html
- https://bun.uptrace.dev/guide/query-update.html
- https://bun.uptrace.dev/guide/query-delete.html
- https://bun.uptrace.dev/guide/query-where.html

Use this chapter when you want the core Bun query operations in one place. It combines the main read and write paths with the safe filtering rules Bun expects you to follow.

## Overview

Bun's query chapters all follow the same idea: stay close to SQL, keep values parameterized, and make the query shape obvious enough that another engineer can review it without guessing. This merged reference covers the most common CRUD and filtering operations together because they are usually used together in real code.

## General Query Rules

- Prefer the Bun builder first.
- Fall back to raw SQL only when the builder cannot express the query clearly.
- Keep values as values through placeholders.
- Keep identifiers controlled and allow-listed.
- Keep each query aligned with one business purpose.
- Scan into the smallest target that matches the result.
- Keep `RETURNING` or scan-back behavior only when the caller needs the row data.

## Select

`NewSelect` is the main read path. Use it for list endpoints, detail lookups, counts, reports, and queries that need joins or subqueries.

### Common Select Tools

- `Model` and `ModelTableExpr` for the row source.
- `Column` and `ColumnExpr` for projection.
- `Table` and `TableExpr` for plain and raw table targets.
- `Join` and `JoinOn` for manual joins.
- `Where`, `WherePK`, and `WhereGroup` for filtering.
- `Group`, `GroupExpr`, `Having`, `Order`, and `OrderExpr` for aggregation and sorting.
- `Limit`, `Offset`, and `For` for paging and locking.
- `Relation` for related rows.
- `Distinct` when duplicate rows should collapse.
- `Count`, `ScanAndCount`, `Exists`, `Rows`, and `ScanRow` for result handling.

### Select Pattern

```go
var users []User
err := db.NewSelect().
	Model(&users).
	Where("status = ?", "active").
	OrderExpr("created_at DESC").
	Limit(20).
	Scan(ctx)
```

### Select Notes

- Use `WherePK()` when the row is already identified.
- Use DTOs for report-style responses.
- Use relation loading when the relationship is part of the model contract.
- Use manual joins when you need tighter SQL control.
- Use raw SQL when it is clearer than the builder.
- Use `Count`, `ScanAndCount`, and `Exists` for common result shapes instead of inventing custom logic.

## Insert

`NewInsert` handles single-row inserts, bulk inserts, returning values, conflict handling, and insert-select style flows.

### Common Insert Shapes

- insert one model,
- insert a slice of models,
- insert with `Returning`,
- upsert with conflict handling,
- insert from a select or a values source.

### Insert Pattern

```go
user := &User{Name: "Alice", Email: "alice@example.com"}

_, err := db.NewInsert().
	Model(user).
	Returning("*").
	Exec(ctx)
```

```go
users := []User{
	{Name: "Bob", Email: "bob@example.com"},
	{Name: "Carol", Email: "carol@example.com"},
}

_, err := db.NewInsert().
	Model(&users).
	Exec(ctx)
```

### Insert Notes

- Bulk insert is the normal way to write a slice of models.
- `Returning("*")` is useful when the database can return generated columns and the caller needs them.
- Upsert should be explicit about the conflict target and update action.
- Insert from a select or values source is useful for staging and backfill jobs.

### Insert Upsert Pattern

```go
_, err := db.NewInsert().
	Model(&book).
	On("CONFLICT (id) DO UPDATE").
	Set("title = EXCLUDED.title").
	Exec(ctx)
```

On MySQL or MariaDB-style flows, use the duplicate-key form that the target backend supports.

## Update

`NewUpdate` is the main update path. Use it for partial updates, explicit column updates, and bulk or table-targeted updates.

### Common Update Tools

- `Set` and `SetColumn` for assignment.
- `WherePK` for primary-key updates.
- `OmitZero` when zero values should be skipped.
- `Table`, `TableExpr`, and `ModelTableExpr` for advanced table targeting.

### Update Pattern

```go
_, err := db.NewUpdate().
	Model(user).
	WherePK().
	OmitZero().
	Exec(ctx)
```

```go
_, err := db.NewUpdate().
	Model(user).
	Column("name", "updated_at").
	WherePK().
	Returning("*").
	Scan(ctx)
```

### Update Notes

- Update only the fields that should change.
- Keep zero-value behavior deliberate.
- Prefer explicit column lists when the intent matters.
- Use `Returning` when the caller needs the updated row back.
- Use `Bulk` when a slice of models needs to be updated with one SQL statement.
- Use `SetColumn` when you need cross-table assignment or backend-specific update shapes.

### Bulk And Multi-Table Updates

The official update guide shows that bulk updates can be built from a slice of models, and multi-table updates can differ between PostgreSQL and MySQL. Bun's role is to keep the SQL shape expressive across those backends.

Use this pattern when:

- many rows need the same change,
- an update depends on joined source rows,
- or you need the database to update rows using an explicit relational source.

## Delete

`NewDelete` handles destructive SQL. Deletion should be explicit because the semantics differ depending on whether the model uses soft delete.

### Delete Pattern

```go
_, err := db.NewDelete().
	Model(&user).
	WherePK().
	Exec(ctx)
```

```go
_, err := db.NewDelete().
	Model((*User)(nil)).
	Where("status = ?", "inactive").
	Exec(ctx)
```

### Delete Notes

- Use soft deletes when recovery or auditability matters.
- Keep hard deletes explicit and rare.
- Make the target rows obvious.
- Be clear about how dependent rows are handled.
- Use `DELETE ... USING`-style shapes when the backend and query shape require a join-based delete.

### Delete Using Joins

The Bun delete guide includes delete-through-join support. That is useful when the rows to delete are identified by another table or derived source. Keep the join condition narrow and reviewable so the destructive query remains obvious.

## Where And Filter Logic

Dynamic filters should stay readable and safe.

### Main Tools

- `Where` for normal predicates.
- `WhereOr` for alternative branches.
- `WhereGroup` for nested boolean logic.
- `WherePK` for primary-key filters.
- `bun.In` for slices.
- `bun.Ident` for controlled identifiers.
- `bun.Safe` only for trusted SQL fragments.

### Filter Rules

- Keep values parameterized.
- Keep identifiers allow-listed.
- Use grouped conditions when OR logic gets complex.
- Never concatenate user input into SQL fragments.
- Use `QueryBuilder` or `ApplyQueryBuilder` when you need to share filter logic across select, update, or delete queries.

### Shared Filter Builder

```go
func addActiveFilter(q bun.QueryBuilder) bun.QueryBuilder {
	return q.Where("status = ?", "active")
}

qb := db.NewSelect().QueryBuilder()
qb = addActiveFilter(qb)
```

The shared-builder approach is useful when the same predicate should apply to several query families without duplicating the logic.

### Filter Example

```go
q := db.NewSelect().Model((*User)(nil))
q = q.Where("status = ?", "active")
q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
	return q.WhereOr("name LIKE ?", "%admin%").
		WhereOr("email LIKE ?", "%admin%")
})
q = q.Where("role IN (?)", bun.In([]string{"admin", "owner"}))
q = q.Where("? = ?", bun.Ident("status"), "active")
```

## Representative Examples

### Read

```go
count, err := db.NewSelect().
	Model((*User)(nil)).
	Where("status = ?", "active").
	Count(ctx)
```

### Insert

```go
_, err := db.NewInsert().
	Model(&user).
	Returning("*").
	Exec(ctx)
```

### Update

```go
_, err := db.NewUpdate().
	Model(&user).
	Column("name", "updated_at").
	WherePK().
	Exec(ctx)
```

### Delete

```go
_, err := db.NewDelete().
	Model((*User)(nil)).
	Where("id = ?", 1).
	Exec(ctx)
```

### Bulk And Advanced Shapes

```go
_, err := db.NewUpdate().
	Model(&books).
	Column("title").
	Bulk().
	Exec(ctx)
```

```go
_, err := db.NewDelete().
	Model((*User)(nil)).
	Where("id IN (?)", bun.In([]int64{1, 2, 3})).
	Exec(ctx)
```

## Practical Guidance

- Use the builder first for inserts, selects, updates, deletes, and filters.
- Use raw SQL only when the builder becomes awkward or unclear.
- Keep query code near the feature or repository that owns it.
- Keep scan targets aligned with the response shape.
- Keep destructive SQL narrow and reviewable.
- Keep dynamic filters allow-listed and parameterized.
- Use `Returning` only when the caller needs values back from the database.
- Prefer `Bulk` and `QueryBuilder` helpers when they make the SQL shape clearer.
