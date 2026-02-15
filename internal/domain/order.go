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

package domain

import (
	"time"

	"github.com/google/uuid"
)

// OrderStatus represents the current state of an order
type OrderStatus string

// Valid order statuses.
const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

// ValidStatuses returns all valid order statuses
func ValidStatuses() []OrderStatus {
	return []OrderStatus{
		OrderStatusPending,
		OrderStatusConfirmed,
		OrderStatusProcessing,
		OrderStatusShipped,
		OrderStatusDelivered,
		OrderStatusCancelled,
	}
}

// CanTransitionTo checks if status transition is valid
func (s OrderStatus) CanTransitionTo(newStatus OrderStatus) bool {
	validTransitions := map[OrderStatus][]OrderStatus{
		OrderStatusPending:    {OrderStatusConfirmed, OrderStatusCancelled},
		OrderStatusConfirmed:  {OrderStatusProcessing, OrderStatusCancelled},
		OrderStatusProcessing: {OrderStatusShipped, OrderStatusCancelled},
		OrderStatusShipped:    {OrderStatusDelivered},
		OrderStatusDelivered:  {},
		OrderStatusCancelled:  {},
	}

	allowed := validTransitions[s]
	for _, status := range allowed {
		if status == newStatus {
			return true
		}
	}
	return false
}

// Order represents a customer order
type Order struct {
	ID         uuid.UUID
	CustomerID string
	Items      []OrderItem
	Status     OrderStatus
	Total      float64
	Version    int // Optimistic locking version, incremented on each update
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

// CalculateTotal computes the total from items
func (o *Order) CalculateTotal() float64 {
	total := 0.0
	for _, item := range o.Items {
		total += item.Subtotal
	}
	return total
}

// Validate performs domain validation
func (o *Order) Validate() error {
	if o.CustomerID == "" {
		return ErrInvalidCustomerID
	}
	if len(o.Items) == 0 {
		return ErrNoItems
	}
	for _, item := range o.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return nil
}
