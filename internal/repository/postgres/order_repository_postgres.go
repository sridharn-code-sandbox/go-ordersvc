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

package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sridharn-code-sandbox/go-ordersvc/internal/domain"
	"github.com/sridharn-code-sandbox/go-ordersvc/internal/repository"
)

// orderRepositoryPostgres implements OrderRepository using PostgreSQL
type orderRepositoryPostgres struct {
	pool *pgxpool.Pool
}

// NewOrderRepository creates a new PostgreSQL order repository
func NewOrderRepository(pool *pgxpool.Pool) repository.OrderRepository {
	return &orderRepositoryPostgres{
		pool: pool,
	}
}

func (r *orderRepositoryPostgres) Create(ctx context.Context, order *domain.Order) error {
	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return err
	}

	// Set initial version
	order.Version = 1

	query := `
		INSERT INTO orders (id, customer_id, items, status, total, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.pool.Exec(ctx, query,
		order.ID,
		order.CustomerID,
		itemsJSON,
		order.Status,
		order.Total,
		order.Version,
		order.CreatedAt,
		order.UpdatedAt,
	)

	return err
}

func (r *orderRepositoryPostgres) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	query := `
		SELECT id, customer_id, items, status, total, version, created_at, updated_at, deleted_at
		FROM orders
		WHERE id = $1 AND deleted_at IS NULL
	`

	var order domain.Order
	var itemsJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&order.ID,
		&order.CustomerID,
		&itemsJSON,
		&order.Status,
		&order.Total,
		&order.Version,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(itemsJSON, &order.Items); err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *orderRepositoryPostgres) Update(ctx context.Context, order *domain.Order) error {
	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return err
	}

	// Optimistic locking: only update if version matches, then increment version
	query := `
		UPDATE orders
		SET customer_id = $1,
		    items = $2,
		    status = $3,
		    total = $4,
		    version = version + 1,
		    updated_at = $5
		WHERE id = $6 AND version = $7 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query,
		order.CustomerID,
		itemsJSON,
		order.Status,
		order.Total,
		time.Now(),
		order.ID,
		order.Version,
	)

	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		// Check if order exists to distinguish between not found and version mismatch
		exists, err := r.orderExists(ctx, order.ID.String())
		if err != nil {
			return err
		}
		if !exists {
			return domain.ErrOrderNotFound
		}
		return domain.ErrConcurrentModification
	}

	// Increment version in the order object to reflect the new state
	order.Version++

	return nil
}

func (r *orderRepositoryPostgres) Delete(ctx context.Context, id string) error {
	// Soft delete - set deleted_at timestamp
	query := `
		UPDATE orders
		SET deleted_at = $1, version = version + 1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrOrderNotFound
	}

	return nil
}

func (r *orderRepositoryPostgres) List(ctx context.Context, opts repository.ListOptions) ([]*domain.Order, int64, error) {
	// Build query with optional status filter
	query := `
		SELECT id, customer_id, items, status, total, version, created_at, updated_at, deleted_at
		FROM orders
		WHERE deleted_at IS NULL
	`
	countQuery := `SELECT COUNT(*) FROM orders WHERE deleted_at IS NULL`

	args := []interface{}{}
	argIndex := 1

	if opts.Status != nil {
		query += ` AND status = $` + string(rune('0'+argIndex))
		countQuery += ` AND status = $1`
		args = append(args, *opts.Status)
		argIndex++
	}

	query += ` ORDER BY created_at DESC LIMIT $` + string(rune('0'+argIndex)) + ` OFFSET $` + string(rune('0'+argIndex+1))
	args = append(args, opts.Limit, opts.Offset)

	// Get total count
	var totalCount int64
	countArgs := args[:len(args)-2] // exclude limit and offset
	if len(countArgs) == 0 {
		countArgs = nil
	}
	err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Get orders
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var order domain.Order
		var itemsJSON []byte

		err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&itemsJSON,
			&order.Status,
			&order.Total,
			&order.Version,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.DeletedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if err := json.Unmarshal(itemsJSON, &order.Items); err != nil {
			return nil, 0, err
		}

		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return orders, totalCount, nil
}

func (r *orderRepositoryPostgres) FindByCustomerID(ctx context.Context, customerID string, opts repository.ListOptions) ([]*domain.Order, int64, error) {
	query := `
		SELECT id, customer_id, items, status, total, version, created_at, updated_at, deleted_at
		FROM orders
		WHERE customer_id = $1 AND deleted_at IS NULL
	`
	countQuery := `SELECT COUNT(*) FROM orders WHERE customer_id = $1 AND deleted_at IS NULL`

	args := []interface{}{customerID}
	argIndex := 2

	if opts.Status != nil {
		query += ` AND status = $` + string(rune('0'+argIndex))
		countQuery += ` AND status = $2`
		args = append(args, *opts.Status)
		argIndex++
	}

	query += ` ORDER BY created_at DESC LIMIT $` + string(rune('0'+argIndex)) + ` OFFSET $` + string(rune('0'+argIndex+1))
	args = append(args, opts.Limit, opts.Offset)

	// Get total count
	var totalCount int64
	countArgs := args[:len(args)-2]
	err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Get orders
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var order domain.Order
		var itemsJSON []byte

		err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&itemsJSON,
			&order.Status,
			&order.Total,
			&order.Version,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.DeletedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if err := json.Unmarshal(itemsJSON, &order.Items); err != nil {
			return nil, 0, err
		}

		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return orders, totalCount, nil
}

// orderExists checks if an order exists (including deleted ones for version conflict detection)
func (r *orderRepositoryPostgres) orderExists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM orders WHERE id = $1)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, id).Scan(&exists)
	return exists, err
}
