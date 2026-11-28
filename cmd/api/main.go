package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/michaelclaudino7/expense-tracker/internal/auth"
	"github.com/michaelclaudino7/expense-tracker/internal/category"
	"github.com/michaelclaudino7/expense-tracker/internal/report"
	"github.com/michaelclaudino7/expense-tracker/internal/transaction"
	"github.com/michaelclaudino7/expense-tracker/internal/user"
)

func main() {
	ctx := context.Background()

	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET must be set and have at least 32 characters")
	}

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

	// Public routes.
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok"))
	})

	tokens := auth.NewTokenManager(jwtSecret, 24*time.Hour)
	authHandler := auth.NewHandler(user.NewRepository(pool), tokens)
	authHandler.RegisterRoutes(mux)

	// Protected routes: every one goes through the JWT middleware.
	authMW := tokens.Middleware

	categoryHandler := category.NewHandler(category.NewRepository(pool))
	categoryHandler.RegisterRoutes(mux, authMW)

	transactionHandler := transaction.NewHandler(transaction.NewRepository(pool))
	transactionHandler.RegisterRoutes(mux, authMW)

	reportHandler := report.NewHandler(report.NewRepository(pool))
	reportHandler.RegisterRoutes(mux, authMW)

	log.Println("API listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
