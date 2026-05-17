---
name: bun-orm
description: Build, review, and refactor Go data-access code with Uptrace Bun using the official Bun Guide chapter structure. Use this skill whenever the user mentions Bun, github.com/uptrace/bun, SQL-first Go ORM work, Bun query building examples, Bun migrations, Bun fixtures, Bun model best practices, Bun transactions, or comparing Bun with GORM/Ent/sqlc/raw SQL.
origin: ECC
---

# Bun ORM Skill

Use this skill to produce Go database code that follows the official Bun Guide at https://bun.uptrace.dev/guide/. Keep the structure and recommendations aligned with the official guide chapters, not ad hoc ORM categories.

## Official Reference Files

Read only the files that match the current task. These files mirror the official Guide chapters:

- `references/query-building-examples.md`: basic CRUD, joins, subqueries, aggregations, grouping, and report query shapes.
- `references/query-crud-and-filters.md`: select, insert, update, delete, and safe dynamic WHERE patterns in one place.
- `references/query-common-table-expressions.md`: `With`, values CTEs, and layered query composition.
- `references/placeholders.md`: SQL placeholders, `bun.Ident`, `bun.Safe`, `bun.In`, and model placeholders.
- `references/installation-configuration.md`: Bun installation, driver selection, PostgreSQL supported drivers, and connection pool tuning.
- `references/migration-and-schema-management.md`: Atlas + Bun schema planning, migration generation, migration locking, up/down functions, and fixtures.
- `references/drivers-and-dialects.md`: supported databases, driver setup, and dialect selection.
- `references/relations.md`: has-one, belongs-to, has-many, polymorphic, and many-to-many relations.
- `references/models.md`: model tags, nullability, soft delete, and naming rules.
- `references/transactions.md`: `bun.Tx`, `RunInTx`, `bun.IDB`, and transaction boundary guidance.
- `references/soft-deletes.md`: soft-delete behavior, `WhereDeleted`, `WhereAllWithDeleted`, and `ForceDelete`.

## Workflow

1. Inspect the repository before changing code: current database provider, migration system, model package, repository/service pattern, and test rules.
2. Identify which official Bun Guide chapter applies and read the matching reference file.
3. Keep Bun SQL-first: prefer Bun's query builder whenever it can express the query clearly; use `Relation` and `With`/CTEs to collapse repeated reads and avoid N+1 round-trips when possible; fall back to raw SQL only when the builder cannot represent the query cleanly or at all.
4. Use the repository's existing architecture. Do not introduce a second data-access pattern unless the user explicitly requests migration.
5. Thread `context.Context` through database methods.
6. Wrap errors with operation context using `%w`; translate not-found and constraint cases at service/API boundaries.
7. Verify with the repository's normal commands. In this repo, prefix shell commands with `rtk`.

## Query Builder Preference

Prefer Bun's query builder for every query shape it can express clearly. Use query builder composition for inserts, selects, updates, deletes, filters, joins, CTEs, relations, counts, and scans before considering handwritten SQL.

## N+1 Avoidance

When a query would otherwise fan out into repeated reads, prefer `Relation` loading or `With`/CTE composition so the data can be fetched in one SQL-shaped query instead of many round-trips.

## Official Docs First

If a query shape, builder method, or Bun feature is not covered by the current references, use the matching official Bun documentation page as the source of truth before adding or changing code. Do not guess a query pattern when the official docs already describe it.

## Chapter-Aligned Guidance

### Key Features

Read `references/query-building-examples.md` when writing CRUD or complex queries. Use `NewInsert`, `NewSelect`, `NewUpdate`, `NewDelete`, `Relation`, subqueries, `ColumnExpr`, `Group`, `Having`, and DTO scan structs according to the query shape. Prefer the query builder first; switch to raw SQL only when the builder cannot express the query clearly enough.

Read `references/query-crud-and-filters.md`, `references/query-common-table-expressions.md`, and `references/placeholders.md` when the task focuses on CRUD queries, safe SQL composition, or filter logic.

### Migration And Schema Management

Read `references/migration-and-schema-management.md` when schema changes or test data are involved. Prefer Atlas for schema diff and migration generation, and Bun for migration execution and fixtures. Avoid hidden production startup schema mutation.

### Installation And Configuration

Read `references/installation-configuration.md` when setting up Bun for a repository, choosing the PostgreSQL driver, or tuning the connection pool.

### Drivers And Dialects

Read `references/drivers-and-dialects.md` when selecting the database driver, the dialect, or the connection string shape.

### Relations

Read `references/relations.md` when defining belongs-to, has-one, has-many, polymorphic, or many-to-many associations.

Use `Relation` loading or `With`-based composition when the query would otherwise fan out into N+1 round-trips.

### Models

Read `references/models.md` when defining table mappings, nullability, timestamps, soft delete columns, and naming rules.

### Transactions

Read `references/transactions.md` when multi-step writes must commit or roll back together. Prefer `RunInTx` for scoped work and `bun.IDB` for reusable persistence helpers.

### Soft Deletes

Read `references/soft-deletes.md` when the model should keep rows instead of physically deleting them. Use `WhereDeleted`, `WhereAllWithDeleted`, and `ForceDelete` deliberately.

## Review Checklist

- The implementation follows the relevant official guide chapter and matching reference file.
- Driver and dialect match the configured database.
- Models use explicit tags where schema stability matters.
- Queries use placeholders for values and avoid user-input string concatenation.
- Complex queries remain readable as SQL-shaped Bun builder code or raw SQL with Bun scanning.
- Transactional helpers use a transaction-compatible shape where needed.
- Migrations and fixtures are explicit when schema or test data changes.
- Tests cover the changed database behavior.
