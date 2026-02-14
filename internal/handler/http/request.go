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

package http

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	CustomerID string      `json:"customer_id" validate:"required,uuid"`
	Items      []OrderItem `json:"items" validate:"required,min=1,dive"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ProductID string  `json:"product_id" validate:"required"`
	Name      string  `json:"name" validate:"required"`
	Quantity  int     `json:"quantity" validate:"required,min=1"`
	Price     float64 `json:"price" validate:"required,gt=0"`
}

// UpdateOrderRequest represents the request to update an order
type UpdateOrderRequest struct {
	Items []OrderItem `json:"items" validate:"required,min=1,dive"`
}

// UpdateStatusRequest represents the request to update order status
type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending confirmed processing shipped delivered cancelled"`
}

// ListOrdersRequest represents the request to list orders
type ListOrdersRequest struct {
	Page     int     `json:"page" validate:"omitempty,min=1"`
	PageSize int     `json:"page_size" validate:"omitempty,min=1,max=100"`
	Status   *string `json:"status" validate:"omitempty,oneof=pending confirmed processing shipped delivered cancelled"`
}
