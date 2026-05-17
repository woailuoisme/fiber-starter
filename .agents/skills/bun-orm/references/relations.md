# Table Relationships

Source: https://bun.uptrace.dev/guide/relations.html

Use this chapter when defining belongs-to, has-one, has-many, polymorphic, or many-to-many relations.

## Overview

The relations chapter is about making joins explicit without losing readability. Bun lets you model relationships in struct tags and then load them when you actually need them.

The key point is that the model declares the relationship shape, while the query decides whether that relationship is loaded. That keeps the SQL readable and prevents unnecessary joins from becoming hidden behavior.

## Relation Types

- `has-one`
- `belongs-to`
- `has-many`
- polymorphic `has-many`
- `many-to-many`
- `m2m:` through an intermediary table

## Loading Pattern

Relation loading is a query-time decision, not a hidden side effect of the model. `Relation("FieldName")` is the normal path when the query needs child rows, and it keeps the SQL readable enough for review.

Use relation loading when:

- the response really needs the related data,
- the relation is part of the model contract,
- and you want Bun to manage the join shape for you.

Use explicit joins when:

- the query needs a custom predicate,
- the relation is not a clean foreign-key path,
- or you want very tight control over projection and aliases.

## Loading Related Rows

`Relation("Name")` tells Bun to load the related data. The guide shows that relation loading should be intentional, especially for list endpoints and larger query graphs. You can also supply a callback to narrow the selected columns on the related table.

Use the callback to prune columns on the related table when the child object is present only for a small slice of fields. That matters for payload size, join cost, and scan time.

You can also influence the join behavior after loading. The official examples show that:

- relation loading is `LEFT JOIN`-like by default,
- you can exclude child columns when you only need the join for filtering,
- and you can force an inner-join-like result by filtering on the relation after loading.

## Has One

`has-one` represents a one-to-one or one-to-zero/one child relationship. Put the relation tag on the field that holds the child row.

Use `has-one` when the child belongs conceptually to the parent and the parent owns the lookup path.

## Belongs To

`belongs-to` represents a parent reference on the child row. This is one of the most common Bun relation shapes because it maps directly to foreign key columns.

Use it when:

- the row stores the foreign key,
- the related row is the parent object,
- and you want Bun to load the parent through the relation tag.

## Has Many

`has-many` represents a slice of child rows. This is the right shape when the parent can own multiple children and you want Bun to load them explicitly.

Use `has-many` when:

- the parent row has many dependent rows,
- the result should be expressed as a slice on the parent model,
- and the query should load the collection only when needed.

## Polymorphic And Many-to-Many

The guide covers polymorphic relations through a virtual type field and many-to-many relations through an intermediary model. The important part is that Bun still wants the model shape and join columns to be explicit.

For many-to-many, register the intermediary model before the first use so Bun can recognize the relation. If the join table carries extra metadata, keep that metadata on the join model rather than trying to infer it from the parent or child model.

For polymorphic relations, keep the type discriminator and the target ID columns obvious in the schema. That keeps the load path understandable when the query is reviewed months later.

## Manual Join Control

If relation loading is not expressive enough, the docs also show manual `Join` and `JoinOn` usage. That is useful when the query needs a specific join condition, a custom alias, or a projection that relation loading would make too broad.

Manual joins are the better choice when:

- the relationship is not a clean foreign-key path,
- the query needs a very specific alias or join predicate,
- the result should be optimized for reporting or export,
- or the SQL should map closely to an existing handwritten query.

## Loading Discipline

- Load relations only when the use case needs them.
- Keep relation names descriptive and stable.
- Prefer explicit relation loading over hidden automatic prefetching.
- Use manual joins when the relation tags would make the query less obvious.
- Register many-to-many intermediary models early.
- Trim relation columns aggressively when the payload does not need the full child model.

## Representative Example

```go
type Post struct {
	bun.BaseModel `bun:"table:posts,alias:p"`

	ID     int64 `bun:"id,pk,autoincrement"`
	UserID int64 `bun:"user_id,notnull"`
	User   *User `bun:"rel:belongs-to,join:user_id=id"`
}

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID    int64   `bun:"id,pk,autoincrement"`
	Posts []*Post `bun:"rel:has-many,join:id=user_id"`
}
```

```go
var posts []Post
err := db.NewSelect().
	Model(&posts).
	Relation("User", func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Column("id", "name")
	}).
	Where("p.id = ?", 1).
	Scan(ctx)
```

```go
func init() {
	db.RegisterModel((*OrderToItem)(nil))
}
```

## Practical Use

- Define relationships explicitly.
- Keep relation aliases stable with `ModelTableExpr`.
- Use relation loading when the query really needs it.
- Use manual joins when you need tighter control over SQL shape.
- Register many-to-many intermediary models early.
- Keep relation queries narrow enough to stay reviewable.
