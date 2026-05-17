# Query Building Examples

Source: https://bun.uptrace.dev/guide/

Use this chapter when you need the main Bun patterns in one place: basic CRUD, joins, subqueries, aggregation, CTEs, relation loading, and scan shapes. This is the reference to reach for first when a query is more than a trivial insert or select.

## Overview

The official guide treats Bun as a SQL-first query builder. The point is not to abstract SQL away, but to help you express SQL clearly from Go. The chapter is useful because it shows the shape Bun expects:

- start from the SQL you would want to run,
- express that shape with Bun's builder,
- keep values parameterized,
- keep identifiers explicit,
- and scan into the smallest target that matches the result.

That means this chapter is both a usage guide and a style guide. If a query is clearer as raw SQL, Bun still lets you do that, but the default path is the builder.

## Basic CRUD Operations

### Create

Creation is normally done with `NewInsert`. Bun supports single-row inserts and bulk inserts with the same overall shape.

```go
user := &User{Name: "Alice", Email: "alice@example.com"}

_, err := db.NewInsert().
	Model(user).
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

Practical notes:

- use the model's tags to control column names and defaults,
- keep bulk inserts for actual batches rather than for one-off writes,
- use returning scans when the database supports it and you need generated fields back,
- and keep inserts close to the repository or service layer that owns the data.

### Read

`NewSelect` is the main query shape for reads. The guide shows the normal flow as selecting from a model, adding filters, and scanning into a slice or a single destination.

```go
var user User
err := db.NewSelect().
	Model(&user).
	Where("id = ?", 1).
	Scan(ctx)
```

```go
var users []User
err := db.NewSelect().
	Model(&users).
	Where("status = ?", "active").
	OrderExpr("created_at DESC").
	Limit(20).
	Scan(ctx)
```

Read patterns you should keep in mind:

- use `WherePK()` when you want the primary key filter,
- use `WhereGroup` when boolean logic gets nested,
- use `OrderExpr` when the ordering is easier to express as raw SQL,
- use DTOs when the result shape is not a table row,
- and use raw queries only when the query builder becomes less clear than handwritten SQL.

### Update

`NewUpdate` is the normal update path. The official examples show both targeted updates and updates that return the changed row.

```go
user.Name = "Alice Cooper"

_, err := db.NewUpdate().
	Model(user).
	Column("name", "updated_at").
	WherePK().
	Exec(ctx)
```

```go
err := db.NewUpdate().
	Model(user).
	Column("name", "updated_at").
	WherePK().
	Returning("*").
	Scan(ctx)
```

Practical notes:

- keep the update narrow by explicitly listing columns,
- use `WherePK()` when the row is already identified,
- keep zero-value behavior aligned with your model tags,
- and prefer small, explicit updates over broad implicit mutation.

### Delete

`NewDelete` handles destructive SQL. The guide's delete examples emphasize that deletion should be explicit, especially when the model uses soft delete.

```go
_, err := db.NewDelete().
	Model((*User)(nil)).
	Where("id = ?", 1).
	Exec(ctx)
```

```go
_, err := db.NewDelete().
	Model(&user).
	WherePK().
	Exec(ctx)
```

If the model uses soft delete, Bun will usually turn a delete into a timestamp update unless you use the hard-delete path. Keep that distinction visible in code review.

## Query Shape Building

The examples chapter is not only about CRUD. It is also about showing how larger SQL should be composed.

The useful pattern is:

1. identify the result shape,
2. decide which table or model owns the base rows,
3. add joins only when needed,
4. add filters in the query, not in memory,
5. keep projection explicit,
6. scan into the smallest target that matches the result.

That is the shape of Bun's query builder across the rest of the guide.

## Joins And Relationships

The query examples show that Bun can use relation loading or explicit joins. Relation loading is good when the model already describes the relationship. Explicit joins are better when you want tighter SQL control.

```go
type Order struct {
	bun.BaseModel `bun:"table:orders,alias:o"`

	ID     int64 `bun:"id,pk,autoincrement"`
	UserID int64 `bun:"user_id"`
	User   User  `bun:"rel:belongs-to,join:user_id=id"`
	Amount int   `bun:"amount"`
}
```

```go
var orders []Order
err := db.NewSelect().
	Model(&orders).
	Relation("User").
	Where("o.amount > ?", 100).
	Scan(ctx)
```

Things to keep in mind:

- use relation loading when the relationship is part of the model contract,
- use `Join` and `JoinOn` when you need more control over the SQL,
- prune child columns when the relation data is large,
- and keep aliases stable so the query remains readable.

## Subqueries

The official query examples use subqueries to build derived filters and layered read logic. This is especially useful when a query needs a set of intermediate rows before producing the final result.

```go
subq := db.NewSelect().
	Model((*User)(nil)).
	Column("id").
	Where("status = ?", "active")

var orders []Order
err := db.NewSelect().
	Model(&orders).
	Where("user_id IN (?)", subq).
	Scan(ctx)
```

Subqueries are a good fit when:

- you need an intermediate set that is not worth materializing in Go,
- the database can do the filtering more efficiently than application code,
- or you want the final SQL to stay readable without repeating the same filter.

## Aggregation And Reporting

The guide's advanced examples show how Bun handles grouped and aggregated queries. This is important for reporting, dashboards, and admin views.

```go
type RevenueRow struct {
	Region string `bun:"region"`
	Total  int64  `bun:"total"`
}

var rows []RevenueRow
err := db.NewSelect().
	TableExpr("orders").
	ColumnExpr("region").
	ColumnExpr("SUM(amount) AS total").
	GroupExpr("region").
	OrderExpr("total DESC").
	Scan(ctx, &rows)
```

Use the following rules for reporting queries:

- scan into report DTOs, not into the main table model,
- keep the aggregation columns named clearly,
- use `Having` for post-aggregation filters,
- use `Group` or `GroupExpr` for stable grouping logic,
- and keep the query reviewable enough that the aggregate shape is obvious.

## CTEs And Layered Composition

CTEs are a key part of Bun's query-building story. They let you name intermediate query steps and then reuse them in the final SQL.

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

CTEs are especially useful for:

- large reporting queries,
- query decomposition,
- reusing a derived table more than once,
- and making bulk flows easier to read.

If a query is already simple, skip the CTE. The point is readability, not extra syntax.

## Values CTEs

The official guide also shows CTEs built from application values. That is useful when you already have a small structured set in memory and want the database to join against it.

```go
type Pair struct {
	ID   int64
	Name string
}

vals := []Pair{
	{ID: 1, Name: "A"},
	{ID: 2, Name: "B"},
}
```

In practice, this pattern helps when:

- you want to update a known set of rows,
- you want to join against a small in-memory dataset,
- or staging into a temporary table would be overkill.

## Scanning Patterns

The query examples section is useful because it shows that Bun is not locked to one scan target. You can scan into:

- a model,
- a slice of models,
- a DTO,
- a scalar,
- a raw map,
- or row-by-row iteration when the result set is large.

```go
var count int
err := db.NewSelect().
	TableExpr("users").
	ColumnExpr("count(*)").
	Scan(ctx, &count)
```

```go
rows, err := db.NewSelect().
	Model((*User)(nil)).
	Rows(ctx)
if err != nil {
	return err
}
defer rows.Close()

for rows.Next() {
	var user User
	if err := db.ScanRow(ctx, rows, &user); err != nil {
		return err
	}
}
```

The main rule is to keep the scan target matched to the query result. Do not force a full table model when the endpoint only needs a report row.

## Query Builder And Raw SQL

The Bun guide is explicit that the builder should help you write SQL, not replace SQL. That gives you a practical rule:

- use the builder first,
- use raw SQL only if the builder gets in the way,
- keep placeholders for values,
- and keep identifiers controlled.

This is the same principle the rest of the query chapters follow.

## Representative Examples

### Basic Select

```go
var users []User
err := db.NewSelect().
	Model(&users).
	Where("status = ?", "active").
	OrderExpr("created_at DESC").
	Limit(20).
	Scan(ctx)
```

### Basic Insert

```go
user := &User{Name: "Alice", Email: "alice@example.com"}

_, err := db.NewInsert().
	Model(user).
	Returning("*").
	Scan(ctx)
```

### Basic Update

```go
_, err := db.NewUpdate().
	Model(&user).
	Column("name", "updated_at").
	WherePK().
	Exec(ctx)
```

### Basic Delete

```go
_, err := db.NewDelete().
	Model((*User)(nil)).
	Where("id = ?", 1).
	Exec(ctx)
```

### Relation Load

```go
var orders []Order
err := db.NewSelect().
	Model(&orders).
	Relation("User").
	Where("o.amount > ?", 100).
	Scan(ctx)
```

### Aggregation

```go
var rows []RevenueRow
err := db.NewSelect().
	TableExpr("orders").
	ColumnExpr("region").
	ColumnExpr("SUM(amount) AS total").
	GroupExpr("region").
	Having("SUM(amount) > ?", 1000).
	Scan(ctx, &rows)
```

### Subquery Filter

```go
activeUsers := db.NewSelect().
	Model((*User)(nil)).
	Column("id").
	Where("status = ?", "active")

err := db.NewSelect().
	Model(&orders).
	Where("user_id IN (?)", activeUsers).
	Scan(ctx)
```

### CTE Composition

```go
regionalSales := db.NewSelect().
	TableExpr("orders").
	ColumnExpr("region").
	ColumnExpr("SUM(amount) AS total_sales").
	GroupExpr("region")

err := db.NewSelect().
	With("regional_sales", regionalSales).
	Table("regional_sales").
	ColumnExpr("region").
	ColumnExpr("total_sales").
	OrderExpr("total_sales DESC").
	Scan(ctx, &rows)
```

## Practical Guidance

- Prefer the Bun builder first, because that keeps SQL readable and composable.
- Use raw SQL only when the builder cannot express the query clearly.
- Keep each query aligned with one business purpose.
- Scan into the smallest shape that answers the use case.
- Use DTOs for reports and list endpoints.
- Keep joins, subqueries, and CTEs explicit enough that the final SQL can still be reviewed.
- Reuse these examples as the template for more specialized query family docs.
