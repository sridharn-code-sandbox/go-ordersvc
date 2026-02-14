// Copyright 2026 go-ordersvc Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/nsridhar76/go-ordersvc/internal/config"
	"github.com/nsridhar76/go-ordersvc/internal/domain"
	httpHandler "github.com/nsridhar76/go-ordersvc/internal/handler/http"
	"github.com/nsridhar76/go-ordersvc/internal/repository"
	"github.com/nsridhar76/go-ordersvc/internal/service"
)

// stubRepository is a temporary stub implementation
type stubRepository struct{}

func (s *stubRepository) Create(ctx context.Context, order *domain.Order) error {
	return nil
}

func (s *stubRepository) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	return nil, domain.ErrOrderNotFound
}

func (s *stubRepository) Update(ctx context.Context, order *domain.Order) error {
	return nil
}

func (s *stubRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (s *stubRepository) List(ctx context.Context, opts repository.ListOptions) ([]*domain.Order, int64, error) {
	return []*domain.Order{}, 0, nil
}

func (s *stubRepository) FindByCustomerID(ctx context.Context, customerID string, opts repository.ListOptions) ([]*domain.Order, int64, error) {
	return []*domain.Order{}, 0, nil
}

// Server holds the HTTP server and its dependencies
type Server struct {
	httpServer *http.Server
	cfg        *config.Config
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config) *Server {
	// TODO: Initialize database connection and repository
	// For now, using mock repository
	repo := &stubRepository{}

	// Create service
	orderService := service.NewOrderService(repo)

	// Create handlers
	orderHandler := httpHandler.NewOrderHandler(orderService)
	healthHandler := httpHandler.NewHealthHandler(cfg.App.Version)

	// Create router
	router := httpHandler.NewRouter(orderHandler, healthHandler)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.HTTPPort),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	return &Server{
		httpServer: httpServer,
		cfg:        cfg,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	fmt.Printf("Starting HTTP server on port %d\n", s.cfg.Server.HTTPPort)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	fmt.Println("Shutting down server...")
	return s.httpServer.Shutdown(ctx)
}

// Run starts the server and handles graceful shutdown
func Run(cfg *config.Config) error {
	server := NewServer(cfg)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	return server.Shutdown(ctx)
}
