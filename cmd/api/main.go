package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/michaelclaudino7/expense-tracker/internal/category"
	"github.com/michaelclaudino7/expense-tracker/internal/report"
	"github.com/michaelclaudino7/expense-tracker/internal/transaction"
)

func main() {
	ctx := context.Background()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://expenses:expenses@localhost:5433/expenses?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to create connection pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to connect to the database: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok"))
	})

	categoryHandler := category.NewHandler(category.NewRepository(pool))
	categoryHandler.RegisterRoutes(mux)

	transactionHandler := transaction.NewHandler(transaction.NewRepository(pool))
	transactionHandler.RegisterRoutes(mux)

	reportHandler := report.NewHandler(report.NewRepository(pool))
	reportHandler.RegisterRoutes(mux)

	log.Println("API listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
