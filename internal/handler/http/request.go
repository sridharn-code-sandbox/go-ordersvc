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

package http //nolint:revive // intentional package name matching handler layer

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	CustomerID string      `json:"customer_id"`
	Items      []OrderItem `json:"items"`
}

// OrderItem represents an item in an order request
type OrderItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// UpdateOrderRequest represents the request to update an order
type UpdateOrderRequest struct {
	Items []OrderItem `json:"items"`
}

// UpdateStatusRequest represents the request to update order status
type UpdateStatusRequest struct {
	Status string `json:"status"`
}
