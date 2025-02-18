# Prerequisites

- Postgres
- Go 1.23.3
- [air](https://github.com/air-verse/air) (optional)

# How to Use

1. Copy the `.env.example` to `.env`
1. Fill in the `.env`
1. Run `go run ./cmd/api/main.go` or just run `air` if you have air installed locally

# Migration

```bash
go run ./cmd/migrator/main.go <Command>
```

| Command | Description                                |
| ------- | ------------------------------------------ |
| create  | Create migration file.                     |
| up      | Migrate existing migrations to the latest. |
| down    | Rollback migration by one file.            |
| status  | See migration status.                      |
| db:seed | Seed the seeder. (For testing purposes)    |
