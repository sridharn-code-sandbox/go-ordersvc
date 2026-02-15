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
	"github.com/redis/go-redis/v9"
)

// rateLimiterRedis implements RateLimiter using Redis
type rateLimiterRedis struct {
	client *redis.Client
}

// NewRateLimiter creates a new Redis rate limiter
func NewRateLimiter(client *redis.Client) cache.RateLimiter {
	return &rateLimiterRedis{
		client: client,
	}
}

func (r *rateLimiterRedis) Allow(_ context.Context, _ string, _ int, _ time.Duration) (bool, error) {
	// TODO: implement
	return true, nil
}

func (r *rateLimiterRedis) Reset(_ context.Context, _ string) error {
	// TODO: implement
	return nil
}
