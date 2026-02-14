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

package service

import (
	"context"

	"github.com/nsridhar76/go-ordersvc/internal/domain"
)

// OrderService defines business logic operations for orders
type OrderService interface {
	// CreateOrder creates a new order with validation
	CreateOrder(ctx context.Context, dto CreateOrderDTO) (*domain.Order, error)

	// GetOrderByID retrieves an order by ID, checking cache first
	GetOrderByID(ctx context.Context, id string) (*domain.Order, error)

	// UpdateOrder updates an existing order
	UpdateOrder(ctx context.Context, id string, dto UpdateOrderDTO) (*domain.Order, error)

	// DeleteOrder soft-deletes an order
	DeleteOrder(ctx context.Context, id string) error

	// ListOrders returns paginated orders with optional status filter
	ListOrders(ctx context.Context, req ListOrdersRequest) (*domain.PaginatedOrders, error)

	// UpdateOrderStatus transitions order to new status with validation
	UpdateOrderStatus(ctx context.Context, id string, newStatus domain.OrderStatus) (*domain.Order, error)
}
