// cmd/api — the COMPOSITION ROOT. The only place that knows the concrete
// wiring: build repos -> inject into services -> inject into the HTTP server.
// Swap repository.NewUserRepo() for a Postgres repo and nothing else changes.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"ecommerce/internal/domain"
	"ecommerce/internal/repository"
	"ecommerce/internal/service"
	transport "ecommerce/internal/transport/http"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// 1. repositories (data layer)
	users := repository.NewUserRepo()
	products := repository.NewProductRepo()
	orders := repository.NewOrderRepo()

	// 2. services (business logic) — injected with repo interfaces
	auth := service.NewAuthService(users)
	productSvc := service.NewProductService(products)
	orderSvc := service.NewOrderService(orders, products)

	// 3. seed an admin so /products can be created out of the box
	if _, err := auth.Register(context.Background(), "admin@shop.com", "admin", "Admin", domain.RoleAdmin); err != nil {
		slog.Error("seed admin failed", "err", err)
		os.Exit(1)
	}
	slog.Info("seeded admin", "email", "admin@shop.com", "password", "admin")

	// 4. transport (HTTP) — injected with services
	srv := transport.NewServer(auth, productSvc, orderSvc)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	slog.Info("listening", "port", port)
	if err := http.ListenAndServe(":"+port, srv.Routes()); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
