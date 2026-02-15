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

// Package repository defines data access interfaces and types.
package repository

import (
	"context"

	"github.com/nsridhar76/go-ordersvc/internal/domain"
)

// OrderRepository defines data access operations for orders
type OrderRepository interface {
	// Create inserts a new order into the database.
	// The order.Version is set to 1 on creation.
	Create(ctx context.Context, order *domain.Order) error

	// FindByID retrieves an order by its ID
	FindByID(ctx context.Context, id string) (*domain.Order, error)

	// Update updates an existing order using optimistic locking.
	// The update will only succeed if the order's version matches the database.
	// On success, the order's version is incremented.
	// Returns domain.ErrConcurrentModification if version mismatch.
	// Returns domain.ErrOrderNotFound if order doesn't exist.
	Update(ctx context.Context, order *domain.Order) error

	// Delete soft-deletes an order by setting deleted_at timestamp.
	// Uses optimistic locking - requires order.Version to match.
	// Returns domain.ErrConcurrentModification if version mismatch.
	Delete(ctx context.Context, id string) error

	// List returns paginated orders with optional status filter
	List(ctx context.Context, opts ListOptions) ([]*domain.Order, int64, error)

	// FindByCustomerID retrieves all orders for a customer
	FindByCustomerID(ctx context.Context, customerID string, opts ListOptions) ([]*domain.Order, int64, error)
}

// ListOptions represents query options for listing orders
type ListOptions struct {
	Limit  int
	Offset int
	Status *domain.OrderStatus
}
