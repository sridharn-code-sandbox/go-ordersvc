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
	"time"

	"github.com/nsridhar76/go-ordersvc/internal/cache"
	"github.com/nsridhar76/go-ordersvc/internal/domain"
	"github.com/redis/go-redis/v9"
)

// orderCacheRedis implements OrderCache using Redis
type orderCacheRedis struct {
	client *redis.Client
}

// NewOrderCache creates a new Redis order cache
func NewOrderCache(client *redis.Client) cache.OrderCache {
	return &orderCacheRedis{
		client: client,
	}
}

func (c *orderCacheRedis) Get(ctx context.Context, id string) (*domain.Order, error) {
	// TODO: implement
	return nil, nil
}

func (c *orderCacheRedis) Set(ctx context.Context, order *domain.Order, ttl time.Duration) error {
	// TODO: implement
	return nil
}

func (c *orderCacheRedis) Delete(ctx context.Context, id string) error {
	// TODO: implement
	return nil
}

func (c *orderCacheRedis) DeletePattern(ctx context.Context, pattern string) error {
	// TODO: implement
	return nil
}
