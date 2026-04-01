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

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/cristianbutiri/property-manager-v2/internal/auth"
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

	jwtSvc := auth.NewJWTService(cfg.JWTSecret)

	exchangeSvc := services.NewExchangeService(cfg.BaseCurrency)
	dashboardSvc := services.NewDashboardService(queries, exchangeSvc)
	importSvc := services.NewImportService(queries)

	syncSvc := services.NewSyncService(queries)
	bgSync := services.NewBackgroundSync(queries, syncSvc, cfg.SyncInterval)
	bgSync.Start()
	defer bgSync.Stop()

	anafSvc := services.NewANAFService()

	var billExtractSvc *services.BillExtractService
	if cfg.AnthropicAPIKey != "" {
		billExtractSvc = services.NewBillExtractService(cfg.AnthropicAPIKey)
		log.Println("bill extraction enabled (ANTHROPIC_API_KEY configured)")
	} else {
		log.Println("bill extraction disabled (ANTHROPIC_API_KEY not configured)")
	}

	var oauthConfig *oauth2.Config
	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		oauthConfig = &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		}
		log.Println("Google OAuth enabled")
	} else {
		log.Println("Google OAuth disabled (GOOGLE_CLIENT_ID and/or GOOGLE_CLIENT_SECRET not configured)")
	}

	authHandler := handlers.NewAuthHandler(queries, jwtSvc, oauthConfig)
	propertyHandler := handlers.NewPropertyHandler(queries)
	bookingHandler := handlers.NewBookingHandler(queries, exchangeSvc)
	expenseHandler := handlers.NewExpenseHandler(queries)
	dashboardHandler := handlers.NewDashboardHandler(dashboardSvc, queries)
	importHandler := handlers.NewImportHandler(importSvc, queries)
	taxRuleHandler := handlers.NewTaxRuleHandler(queries)
	billExtractHandler := handlers.NewBillExtractHandler(billExtractSvc, queries)
	integrationHandler := handlers.NewIntegrationHandler(queries, syncSvc)
	propertyMappingHandler := handlers.NewPropertyMappingHandler(queries)
	sellerProfileHandler := handlers.NewSellerProfileHandler(queries)
	inventoryHandler := handlers.NewInventoryHandler(queries)
	invoiceHandler := handlers.NewInvoiceHandler(queries, anafSvc)

	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.CORS())
	r.Use(chimiddleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		// Public auth routes
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)
		r.Get("/auth/google", authHandler.GoogleLogin)
		r.Get("/auth/google/callback", authHandler.GoogleCallback)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(jwtSvc))

			r.Get("/auth/me", authHandler.Me)

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
					r.Post("/expenses/extract", billExtractHandler.Extract)

					r.Post("/import/csv", importHandler.ImportFile)

					r.Get("/tax-rules", taxRuleHandler.ListByProperty)
					r.Post("/tax-rules", taxRuleHandler.Create)

					r.Get("/mappings", propertyMappingHandler.ListByProperty)
					r.Post("/mappings", propertyMappingHandler.Create)
					r.Delete("/mappings/{mappingId}", propertyMappingHandler.Delete)

					r.Get("/inventory", inventoryHandler.ListByProperty)
					r.Post("/inventory", inventoryHandler.Create)

					r.Get("/dashboard", dashboardHandler.PropertyDashboard)
				})
			})

			r.Delete("/bookings/{id}", bookingHandler.Delete)
			r.Post("/bookings/{id}/invoice", invoiceHandler.GenerateInvoice)
			r.Get("/bookings/{id}/invoice", invoiceHandler.GetBookingInvoice)

			r.Put("/expenses/{id}", expenseHandler.Update)
			r.Delete("/expenses/{id}", expenseHandler.Delete)

			r.Put("/inventory/{id}", inventoryHandler.Update)
			r.Delete("/inventory/{id}", inventoryHandler.Delete)

			r.Put("/tax-rules/{id}", taxRuleHandler.Update)
			r.Delete("/tax-rules/{id}", taxRuleHandler.Delete)

			r.Route("/integrations", func(r chi.Router) {
				r.Get("/", integrationHandler.List)
				r.Post("/", integrationHandler.Create)
				r.Put("/{id}", integrationHandler.Update)
				r.Delete("/{id}", integrationHandler.Delete)
				r.Post("/{id}/sync", integrationHandler.Sync)
			})

			r.Get("/seller-profile", sellerProfileHandler.Get)
			r.Put("/seller-profile", sellerProfileHandler.Upsert)

			r.Post("/anaf/lookup", invoiceHandler.ANAFLookup)

			r.Get("/invoices", invoiceHandler.List)
			r.Get("/invoices/{id}/xml", invoiceHandler.DownloadXML)
			r.Delete("/invoices/{id}", invoiceHandler.Delete)

			r.Get("/dashboard", dashboardHandler.OverallDashboard)
		})
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
