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

// Package domain contains core business entities and value objects.
package domain

import "errors"

// Domain errors for order operations.
var (
	ErrOrderNotFound          = errors.New("order not found")
	ErrInvalidCustomerID      = errors.New("invalid customer ID")
	ErrNoItems                = errors.New("order must have at least one item")
	ErrInvalidProductID       = errors.New("invalid product ID")
	ErrInvalidProductName     = errors.New("invalid product name")
	ErrInvalidQuantity        = errors.New("quantity must be greater than 0")
	ErrInvalidPrice           = errors.New("price must be greater than 0")
	ErrInvalidStatus          = errors.New("invalid order status")
	ErrInvalidTransition      = errors.New("invalid status transition")
	ErrOrderAlreadyDeleted    = errors.New("order is already deleted")
	ErrConcurrentModification = errors.New("order was modified by another process")
)
