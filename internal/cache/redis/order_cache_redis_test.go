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

package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/nsridhar76/go-ordersvc/internal/domain"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMiniredis(t *testing.T) (*miniredis.Miniredis, *goredis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{
		Addr: mr.Addr(),
	})
	return mr, client
}

func newTestOrder() *domain.Order {
	return &domain.Order{
		ID:         uuid.New(),
		CustomerID: "customer-123",
		Items: []domain.OrderItem{
			{
				ID:        uuid.New(),
				ProductID: "product-1",
				Name:      "Test Product",
				Quantity:  2,
				Price:     10.50,
				Subtotal:  21.00,
			},
		},
		Status:    domain.OrderStatusPending,
		Total:     21.00,
		Version:   1,
		CreatedAt: time.Now().Truncate(time.Second),
		UpdatedAt: time.Now().Truncate(time.Second),
	}
}

func TestOrderCacheRedis_SetThenGet_ReturnsOrder(t *testing.T) {
	_, client := setupMiniredis(t)
	cache := NewOrderCache(client)
	ctx := context.Background()
	order := newTestOrder()

	err := cache.Set(ctx, order, 5*time.Minute)
	require.NoError(t, err)

	got, err := cache.Get(ctx, order.ID.String())
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, order.ID, got.ID)
	assert.Equal(t, order.CustomerID, got.CustomerID)
	assert.Equal(t, order.Status, got.Status)
	assert.Equal(t, order.Total, got.Total)
	assert.Equal(t, len(order.Items), len(got.Items))
}

func TestOrderCacheRedis_Get_CacheMiss_ReturnsNil(t *testing.T) {
	_, client := setupMiniredis(t)
	cache := NewOrderCache(client)
	ctx := context.Background()

	got, err := cache.Get(ctx, uuid.New().String())

	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestOrderCacheRedis_Set_WithTTL_Expires(t *testing.T) {
	mr, client := setupMiniredis(t)
	cache := NewOrderCache(client)
	ctx := context.Background()
	order := newTestOrder()

	err := cache.Set(ctx, order, 1*time.Second)
	require.NoError(t, err)

	// Verify it exists
	got, err := cache.Get(ctx, order.ID.String())
	require.NoError(t, err)
	assert.NotNil(t, got)

	// Fast-forward time past TTL
	mr.FastForward(2 * time.Second)

	// Should be expired
	got, err = cache.Get(ctx, order.ID.String())
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestOrderCacheRedis_Delete_RemovesFromCache(t *testing.T) {
	_, client := setupMiniredis(t)
	cache := NewOrderCache(client)
	ctx := context.Background()
	order := newTestOrder()

	err := cache.Set(ctx, order, 5*time.Minute)
	require.NoError(t, err)

	err = cache.Delete(ctx, order.ID.String())
	require.NoError(t, err)

	got, err := cache.Get(ctx, order.ID.String())
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestOrderCacheRedis_Get_InvalidJSON_ReturnsError(t *testing.T) {
	_, client := setupMiniredis(t)
	cache := NewOrderCache(client)
	ctx := context.Background()

	// Write invalid JSON directly
	id := uuid.New().String()
	key := "order:" + id
	client.Set(ctx, key, "not-valid-json", 5*time.Minute)

	got, err := cache.Get(ctx, id)

	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "cache unmarshal")
}
