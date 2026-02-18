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

// Package cache defines caching interfaces for the order service.
package cache

import (
	"context"
	"time"

	"github.com/sridharn-code-sandbox/go-ordersvc/internal/domain"
)

// OrderCache defines caching operations for orders
type OrderCache interface {
	// Get retrieves an order from cache
	Get(ctx context.Context, id string) (*domain.Order, error)

	// Set stores an order in cache with TTL
	Set(ctx context.Context, order *domain.Order, ttl time.Duration) error

	// Delete removes an order from cache
	Delete(ctx context.Context, id string) error

	// DeletePattern removes all keys matching pattern (e.g., "order:customer:123:*")
	DeletePattern(ctx context.Context, pattern string) error
}

// RateLimiter defines rate limiting operations
type RateLimiter interface {
	// Allow checks if request is allowed under rate limit
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)

	// Reset clears rate limit counter for a key
	Reset(ctx context.Context, key string) error
}
