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
	"encoding/json"
	"fmt"
	"time"

	"github.com/sridharn-code-sandbox/go-ordersvc/internal/cache"
	"github.com/sridharn-code-sandbox/go-ordersvc/internal/domain"
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
	key := orderKey(id)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cache get %s: %w", key, err)
	}

	var order domain.Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, fmt.Errorf("cache unmarshal %s: %w", key, err)
	}
	return &order, nil
}

func (c *orderCacheRedis) Set(ctx context.Context, order *domain.Order, ttl time.Duration) error {
	key := orderKey(order.ID.String())
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("cache marshal %s: %w", key, err)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("cache set %s: %w", key, err)
	}
	return nil
}

func (c *orderCacheRedis) Delete(ctx context.Context, id string) error {
	key := orderKey(id)
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("cache del %s: %w", key, err)
	}
	return nil
}

func (c *orderCacheRedis) DeletePattern(ctx context.Context, pattern string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("cache scan %s: %w", pattern, err)
		}
		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("cache del pattern %s: %w", pattern, err)
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

func orderKey(id string) string {
	return "order:" + id
}
