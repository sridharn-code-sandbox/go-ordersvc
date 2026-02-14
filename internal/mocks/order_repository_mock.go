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

package mocks

import (
	"context"

	"github.com/nsridhar76/go-ordersvc/internal/domain"
	"github.com/nsridhar76/go-ordersvc/internal/repository"
)

// OrderRepositoryMock is a mock implementation of OrderRepository
type OrderRepositoryMock struct {
	CreateFunc           func(ctx context.Context, order *domain.Order) error
	FindByIDFunc         func(ctx context.Context, id string) (*domain.Order, error)
	UpdateFunc           func(ctx context.Context, order *domain.Order) error
	DeleteFunc           func(ctx context.Context, id string) error
	ListFunc             func(ctx context.Context, opts repository.ListOptions) ([]*domain.Order, int64, error)
	FindByCustomerIDFunc func(ctx context.Context, customerID string, opts repository.ListOptions) ([]*domain.Order, int64, error)
}

func (m *OrderRepositoryMock) Create(ctx context.Context, order *domain.Order) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, order)
	}
	return nil
}

func (m *OrderRepositoryMock) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *OrderRepositoryMock) Update(ctx context.Context, order *domain.Order) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, order)
	}
	return nil
}

func (m *OrderRepositoryMock) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *OrderRepositoryMock) List(ctx context.Context, opts repository.ListOptions) ([]*domain.Order, int64, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, opts)
	}
	return nil, 0, nil
}

func (m *OrderRepositoryMock) FindByCustomerID(ctx context.Context, customerID string, opts repository.ListOptions) ([]*domain.Order, int64, error) {
	if m.FindByCustomerIDFunc != nil {
		return m.FindByCustomerIDFunc(ctx, customerID, opts)
	}
	return nil, 0, nil
}
