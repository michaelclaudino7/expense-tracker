# Expense Tracker

A personal expense tracking REST API built with Go and PostgreSQL.

## Features

- Categories for income and expenses
- Transactions with amounts stored in cents (no floating point errors)
- Transaction filters by date range and category
- Monthly report with totals per category and balance

## Tech stack

- Go 1.22+ (standard library `net/http` router)
- PostgreSQL 16 (Docker)
- pgx / pgxpool
- golang-migrate

## Getting started

1. Start the database: `docker compose up -d`
2. Apply migrations: `migrate -path migrations -database "postgres://expenses:expenses@localhost:5433/expenses?sslmode=disable" up`
3. Run the API: `go run ./cmd/api`
4. Check it: `curl.exe localhost:8080/health`

The database listens on port 5433 on the host to avoid clashing with a local PostgreSQL on 5432.

## Endpoints

| Method | Route | Description |
|---|---|---|
| GET | /health | Health check |
| POST | /categories | Create a category |
| GET | /categories | List categories |
| GET | /categories/{id} | Get a category |
| PUT | /categories/{id} | Update a category |
| DELETE | /categories/{id} | Delete a category |
| POST | /transactions | Create a transaction |
| GET | /transactions | List transactions (optional `from`, `to`, `category_id`) |
| GET | /transactions/{id} | Get a transaction |
| PUT | /transactions/{id} | Update a transaction |
| DELETE | /transactions/{id} | Delete a transaction |
| GET | /reports/monthly?month=YYYY-MM | Monthly report |

## Roadmap

- [x] Unit tests for input validation
- [ ] Pagination
- [ ] User accounts and JWT authentication
- [ ] Budget per category
- [ ] CSV export