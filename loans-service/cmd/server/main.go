package main

import (
"context"
"database/sql"
"encoding/json"
"log"
"net/http"
"time"

"github.com/go-chi/chi/v5"

"github.com/jdeleonchang/library-technical-test/loans-service/internal/loans"
"github.com/jdeleonchang/library-technical-test/loans-service/internal/platform/database"
"github.com/jdeleonchang/library-technical-test/loans-service/internal/platform/httpclient"
)

type HealthResponse struct {
Status   string `json:"status"`
Service  string `json:"service"`
Database string `json:"database"`
}

func main() {
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

db, err := database.Open(ctx)
if err != nil {
log.Fatalf("database connection failed: %v", err)
}
defer db.Close()

if err := database.RunMigrations(ctx, db); err != nil {
log.Fatalf("database migration failed: %v", err)
}

loanRepository := loans.NewRepository(db)
libraryClient := httpclient.NewLibraryClient()
loanService := loans.NewService(loanRepository, libraryClient)
loanHandler := loans.NewHandler(loanService)

router := chi.NewRouter()

router.Get("/health", healthHandler(db))
loanHandler.RegisterRoutes(router)

log.Println("Loans service running on port 8080")

err = http.ListenAndServe(":8080", router)
if err != nil {
log.Fatal(err)
}
}

func healthHandler(db *sql.DB) http.HandlerFunc {
return func(w http.ResponseWriter, r *http.Request) {
ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
defer cancel()

databaseStatus := "ok"
if err := db.PingContext(ctx); err != nil {
databaseStatus = "unavailable"
w.WriteHeader(http.StatusServiceUnavailable)
} else {
w.WriteHeader(http.StatusOK)
}

w.Header().Set("Content-Type", "application/json")

response := HealthResponse{
Status:   "ok",
Service:  "loans-service",
Database: databaseStatus,
}

if err := json.NewEncoder(w).Encode(response); err != nil {
log.Printf("failed to encode health response: %v", err)
}
}
}
