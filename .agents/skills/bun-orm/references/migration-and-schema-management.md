# Migration And Schema Management

Source:
- https://bun.uptrace.dev/guide/migrations.html
- https://atlasgo.io/guides/orms/bun

Use this chapter when adding schema changes, generating migration files, running migrations, or loading fixtures. In this project, prefer Atlas for schema planning and migration generation, and use Bun for the ORM and runtime database access.

## Overview

Bun already provides migration execution, rollback handling, and fixtures. Atlas complements that by reading Bun models, calculating schema diffs, and generating the SQL migration files that Bun can execute.

The practical workflow is:

- define the desired schema in Bun models,
- ask Atlas to diff the models against the current database,
- review the generated SQL,
- then apply the migration using your Bun migration workflow.

That split gives you automatic planning with Atlas and robust execution with Bun.

## Core Concepts

- migrations can be written in Go or SQL,
- `bun_migrations` stores applied migration state,
- `migrate.NewMigrator` creates the migrator,
- migration execution can be locked to avoid concurrent runs,
- Atlas can generate migration SQL from Bun model changes,
- Atlas can detect schema drift between models and production databases,
- the generated SQL can be used inside Bun's migration system.

## Atlas + Bun Workflow

Atlas is the schema planner. Bun is the ORM and migration executor.

Use Atlas when you want:

- migration diffs generated from model changes,
- drift detection against production,
- migration linting before deployment,
- and a schema source of truth that follows your Bun models.

Use Bun when you want:

- the runtime migrator,
- the migration history table,
- rollback execution,
- fixtures,
- and normal application database access.

## Atlas Provider Modes

The Atlas Bun provider supports two modes.

### Standalone Mode

Use standalone mode when all Bun models live in a single package and the provider can load them directly.

Typical setup:

- create `atlas.hcl`,
- point Atlas at the model package,
- set the desired dialect,
- and run `atlas migrate diff`.

### Go Program Mode

Use Go program mode when your Bun model loading requires more control.

This mode is also required for more advanced cases such as many-to-many relation loading, where the join table must be registered explicitly.

Typical setup:

- create a small Go loader program,
- use the Atlas Bun provider from that program,
- load the model types you want to diff,
- and then run Atlas against the loader output.

## Atlas Installation

Atlas can be installed through the Atlas installer script, and the Bun provider is installed as a Go dependency.

```bash
curl -sSf https://atlasgo.sh | sh
go get -u ariga.io/atlas-provider-bun
```

If you use the provider in a Go program, keep the tool dependency pinned in the repo so Go modules do not drop it during tidy operations.

## Atlas Configuration

Atlas uses `atlas.hcl` to describe the desired schema source, dev database, migration directory, and diff formatting.

The important settings are:

- the external schema program that loads the Bun models,
- the dev database used for diff planning,
- the migration directory,
- and the SQL formatting for generated files.

## Example Atlas Configuration

```hcl
data "external_schema" "bun" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-bun",
    "load",
    "--path", "./path/to/models",
    "--dialect", "mysql",
  ]
}

env "bun" {
  src = data.external_schema.bun.url
  dev = "docker://mysql/8/dev"
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
```

This pattern lets Atlas inspect your Bun models and emit SQL migration files into the migration directory you already use.

## Recommended Atlas Commands

The practical commands you usually need are:

- `atlas migrate diff` to generate a migration from model changes,
- `atlas migrate lint` to check the generated or existing migrations,
- `atlas migrate apply` when you want Atlas to apply migrations directly in a non-Bun flow,
- and the Bun migrator when your application owns migration execution.

In this project, the preferred split is:

- Atlas generates and validates the migration SQL,
- Bun executes the application migration path and keeps the migration history consistent.

### Typical Local Workflow

1. Change the Bun models.
2. Run an Atlas diff against the Bun schema.
3. Review the generated SQL.
4. Commit the migration output.
5. Apply the migration through the repo's chosen execution flow.

### Typical CI Workflow

1. Run Atlas lint against the current schema files.
2. Compare the Bun models and the expected schema.
3. Fail the build if there is drift or a risky migration change.
4. Keep the migration files under version control so reviews are repeatable.

## Many-to-Many Relations

Atlas supports Bun many-to-many relations, but the provider must be used carefully.

- In standalone mode, join-table discovery is limited.
- In Go program mode, register the join table explicitly.
- Pass the join table model together with the related models to the loader.

That matters because join-table relationships are a schema concern, not just a query concern.

## Bun Migration Execution

Bun still remains useful for execution.

- Go migrations are useful when the migration needs custom logic.
- SQL migrations are useful when the generated SQL should be the source of truth.
- Transactional SQL files can use Bun's SQL migration conventions.
- The migrator still manages lock, status, rollback, and execution flow.

Atlas and Bun fit together well when Atlas generates the migration body and Bun manages the application-side execution.

## Choosing The Execution Path

Use Bun execution when:

- the app already owns the migration runner,
- you want the migration history table managed from Go,
- or you want Bun fixtures and migrator helpers in the same codebase.

Use Atlas execution when:

- you want Atlas to manage the apply step in an external workflow,
- you are operating in a pure schema-management pipeline,
- or the deployment tooling already standardizes on Atlas commands.

## Go Migrations

The guide shows registering up/down functions in a migration collection. That is useful when schema changes are simple enough to express in code and when you want to keep the migration logic in Go.

Go migrations are a good fit when:

- the migration needs conditional logic,
- the change touches both schema and seed data,
- or the application already owns the domain logic needed for the transformation.

## SQL Migrations

SQL files are also supported and are useful when you want the SQL itself to be the source of truth. The guide points out that `.tx.up.sql` can be used for transactional SQL migrations.

The SQL migration path is a strong choice when:

- the migration is easier to review as raw SQL,
- you want the exact DDL to be visible,
- or the database-specific syntax matters.

Atlas-generated migrations are a natural fit here because Atlas produces the SQL and Bun executes it.

## Fixtures

Fixtures are part of the official Bun story. They are useful for deterministic test data and for initial data loading in development or CI.

That makes them a good fit for integration tests and repeatable local setup, especially when tests need a stable database baseline.

## Rollback Behavior

The docs warn that migration rollback behavior is grouped and should be planned carefully. That means production migration strategy needs to be deliberate rather than casual.

The important implication is that the rollback unit is not necessarily a single file. Treat a migration batch as the unit you are willing to revert together.

## Drift Detection And Review

Atlas adds value when you want to catch schema drift before it reaches production.

The useful workflow is:

- change Bun models,
- generate a diff with Atlas,
- review the generated SQL,
- apply the migration,
- and keep the migration output under version control.

That makes the schema change visible and reviewable instead of implicit.

## Review Rules

- Review the generated SQL before applying it.
- Keep the migration paired with the model change that produced it.
- Do not hide schema changes in application startup logic.
- Use the migration diff as a review artifact, not just a generated file.

## Representative Example

```go
migrator := migrate.NewMigrator(db, migrations.New())
if err := migrator.Lock(ctx); err != nil {
	return err
}
defer migrator.Unlock(ctx)

_, err := migrator.Migrate(ctx)
```

```sql
-- Example SQL migration with a transactional split marker.
SELECT 1 --bun:split
SELECT 2
```

## Practical Use

- Prefer Atlas for migration planning and SQL generation.
- Use Bun for migration execution, rollback, and fixtures.
- Keep migration naming stable and chronological.
- Review schema migrations together with model tag changes and query code.
- Run Atlas diff before committing schema changes.
- Keep generated SQL in version control so drift is visible.
