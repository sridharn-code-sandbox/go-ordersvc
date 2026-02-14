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

import "github.com/google/uuid"

// OrderItem represents a single item in an order
type OrderItem struct {
	ID        uuid.UUID
	ProductID string
	Name      string
	Quantity  int
	Price     float64
	Subtotal  float64
}

// CalculateSubtotal computes item subtotal
func (i *OrderItem) CalculateSubtotal() float64 {
	return float64(i.Quantity) * i.Price
}

// Validate performs item validation
func (i *OrderItem) Validate() error {
	if i.ProductID == "" {
		return ErrInvalidProductID
	}
	if i.Name == "" {
		return ErrInvalidProductName
	}
	if i.Quantity <= 0 {
		return ErrInvalidQuantity
	}
	if i.Price <= 0 {
		return ErrInvalidPrice
	}
	return nil
}
