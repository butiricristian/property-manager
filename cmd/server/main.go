package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cristianbutiri/property-manager-v2/internal/config"
	"github.com/cristianbutiri/property-manager-v2/internal/database/sqlc"
	"github.com/cristianbutiri/property-manager-v2/internal/handlers"
	"github.com/cristianbutiri/property-manager-v2/internal/middleware"
	"github.com/cristianbutiri/property-manager-v2/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("connected to database")

	queries := sqlc.New(pool)

	dashboardSvc := services.NewDashboardService(queries)
	importSvc := services.NewImportService(queries)

	propertyHandler := handlers.NewPropertyHandler(queries)
	bookingHandler := handlers.NewBookingHandler(queries)
	expenseHandler := handlers.NewExpenseHandler(queries)
	dashboardHandler := handlers.NewDashboardHandler(dashboardSvc)
	importHandler := handlers.NewImportHandler(importSvc)

	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.CORS())
	r.Use(chimiddleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Route("/properties", func(r chi.Router) {
			r.Get("/", propertyHandler.List)
			r.Post("/", propertyHandler.Create)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", propertyHandler.Get)
				r.Put("/", propertyHandler.Update)
				r.Delete("/", propertyHandler.Delete)

				r.Get("/bookings", bookingHandler.ListByProperty)
				r.Post("/bookings", bookingHandler.Create)

				r.Get("/expenses", expenseHandler.ListByProperty)
				r.Post("/expenses", expenseHandler.Create)

				r.Post("/import/csv", importHandler.ImportFile)

				r.Get("/dashboard", dashboardHandler.PropertyDashboard)
			})
		})

		r.Delete("/bookings/{id}", bookingHandler.Delete)

		r.Put("/expenses/{id}", expenseHandler.Update)
		r.Delete("/expenses/{id}", expenseHandler.Delete)

		r.Get("/dashboard", dashboardHandler.OverallDashboard)
	})

	// Serve frontend static files
	handlers.ServeStaticFiles(r)

	srv := &http.Server{
		Addr:    cfg.Addr(),
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}()

	log.Printf("server starting on %s", cfg.Addr())
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
	log.Println("server stopped")
}
