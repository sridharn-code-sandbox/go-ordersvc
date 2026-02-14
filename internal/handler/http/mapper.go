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

import (
	"github.com/nsridhar76/go-ordersvc/internal/domain"
)

// MapOrderToResponse maps a domain order to HTTP response
func MapOrderToResponse(order *domain.Order) OrderResponse {
	items := make([]OrderItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = OrderItem{
			ProductID: item.ProductID,
			Name:      item.Name,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	return OrderResponse{
		ID:         order.ID.String(),
		CustomerID: order.CustomerID,
		Items:      items,
		Status:     string(order.Status),
		Total:      order.Total,
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
	}
}

// MapRequestToOrderItems maps HTTP request items to domain items
func MapRequestToOrderItems(items []OrderItem) []domain.OrderItem {
	domainItems := make([]domain.OrderItem, len(items))
	for i, item := range items {
		domainItems[i] = domain.OrderItem{
			ProductID: item.ProductID,
			Name:      item.Name,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}
	return domainItems
}
