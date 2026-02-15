# API Reference

This document describes the REST API for the go-ordersvc order management service.

**Base URL:** `/api/v1`

## Authentication

Currently no authentication is required. Health endpoints (`/healthz`, `/readyz`) are always unauthenticated for Kubernetes probe compatibility.

---

## Orders

### Create Order

Creates a new order.

**Endpoint:** `POST /api/v1/orders`

**Request Body:**

```json
{
  "customer_id": "string (required)",
  "items": [
    {
      "product_id": "string",
      "name": "string",
      "quantity": 1,
      "price": 29.99
    }
  ]
}
```

**Response:** `201 Created`

**Headers:**
- `Location: /api/v1/orders/{id}` - URL of the created resource

**Response Body:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "customer_id": "cust-123",
  "items": [
    {
      "id": "item-uuid",
      "product_id": "prod-1",
      "name": "Product Name",
      "quantity": 2,
      "price": 29.99,
      "subtotal": 59.98
    }
  ],
  "status": "pending",
  "total": 59.98,
  "version": 1,
  "created_at": "2026-02-14T12:00:00Z",
  "updated_at": "2026-02-14T12:00:00Z"
}
```

**Error Responses:**

| Status | Code | Description |
|--------|------|-------------|
| 400 | `MISSING_CUSTOMER_ID` | customer_id field is empty |
| 400 | `MISSING_ITEMS` | items array is empty |
| 400 | `INVALID_REQUEST` | Malformed JSON body |
| 500 | `INTERNAL_ERROR` | Server error |

**Example:**

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "cust-123",
    "items": [
      {"product_id": "prod-1", "name": "Widget", "quantity": 2, "price": 29.99}
    ]
  }'
```

---

### Get Order

Retrieves a single order by ID.

**Endpoint:** `GET /api/v1/orders/{id}`

**Path Parameters:**

| Name | Type | Description |
|------|------|-------------|
| id | uuid | Order ID |

**Response:** `200 OK`

**Response Body:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "customer_id": "cust-123",
  "items": [...],
  "status": "pending",
  "total": 59.98,
  "version": 1,
  "created_at": "2026-02-14T12:00:00Z",
  "updated_at": "2026-02-14T12:00:00Z"
}
```

**Error Responses:**

| Status | Code | Description |
|--------|------|-------------|
| 400 | `MISSING_ID` | No ID provided |
| 404 | `ORDER_NOT_FOUND` | Order does not exist |
| 500 | `INTERNAL_ERROR` | Server error |

**Example:**

```bash
curl http://localhost:8080/api/v1/orders/550e8400-e29b-41d4-a716-446655440000
```

---

### List Orders

Retrieves a paginated list of orders with optional status filtering.

**Endpoint:** `GET /api/v1/orders`

**Query Parameters:**

| Name | Type | Default | Max | Description |
|------|------|---------|-----|-------------|
| limit | int | 20 | 100 | Items per page |
| offset | int | 0 | - | Pagination offset |
| status | string | - | - | Filter by status |

**Valid status values:** `pending`, `confirmed`, `processing`, `shipped`, `delivered`, `cancelled`

**Response:** `200 OK`

**Response Body:**

```json
{
  "orders": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "customer_id": "cust-123",
      "items": [...],
      "status": "pending",
      "total": 59.98,
      "version": 1,
      "created_at": "2026-02-14T12:00:00Z",
      "updated_at": "2026-02-14T12:00:00Z"
    }
  ],
  "total": 100,
  "limit": 20,
  "offset": 0
}
```

**Example:**

```bash
# List first 20 orders
curl "http://localhost:8080/api/v1/orders"

# List pending orders with pagination
curl "http://localhost:8080/api/v1/orders?status=pending&limit=10&offset=20"
```

---

### Update Order

Updates an existing order's items.

**Endpoint:** `PUT /api/v1/orders/{id}`

**Path Parameters:**

| Name | Type | Description |
|------|------|-------------|
| id | uuid | Order ID |

**Request Body:**

```json
{
  "items": [
    {
      "product_id": "prod-2",
      "name": "New Product",
      "quantity": 3,
      "price": 19.99
    }
  ]
}
```

**Response:** `200 OK`

**Response Body:** Updated order object

**Error Responses:**

| Status | Code | Description |
|--------|------|-------------|
| 400 | `MISSING_ID` | No ID provided |
| 400 | `MISSING_ITEMS` | items array is empty |
| 400 | `INVALID_REQUEST` | Malformed JSON body |
| 404 | `ORDER_NOT_FOUND` | Order does not exist |
| 409 | `CONCURRENT_MODIFICATION` | Optimistic lock conflict |
| 500 | `INTERNAL_ERROR` | Server error |

**Example:**

```bash
curl -X PUT http://localhost:8080/api/v1/orders/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {"product_id": "prod-2", "name": "New Widget", "quantity": 1, "price": 49.99}
    ]
  }'
```

---

### Update Order Status

Updates an order's status. Only valid state transitions are allowed.

**Endpoint:** `PATCH /api/v1/orders/{id}/status`

**Path Parameters:**

| Name | Type | Description |
|------|------|-------------|
| id | uuid | Order ID |

**Request Body:**

```json
{
  "status": "confirmed"
}
```

**Valid Status Transitions:**

| From | Allowed Transitions |
|------|---------------------|
| pending | confirmed, cancelled |
| confirmed | processing, cancelled |
| processing | shipped, cancelled |
| shipped | delivered |
| delivered | (terminal state) |
| cancelled | (terminal state) |

**Response:** `200 OK`

**Response Body:** Updated order object with incremented version

**Error Responses:**

| Status | Code | Description |
|--------|------|-------------|
| 400 | `MISSING_ID` | No ID provided |
| 400 | `MISSING_STATUS` | status field is empty |
| 400 | `INVALID_TRANSITION` | Status transition not allowed |
| 404 | `ORDER_NOT_FOUND` | Order does not exist |
| 409 | `CONCURRENT_MODIFICATION` | Optimistic lock conflict |
| 500 | `INTERNAL_ERROR` | Server error |

**Example:**

```bash
# Confirm a pending order
curl -X PATCH http://localhost:8080/api/v1/orders/550e8400-e29b-41d4-a716-446655440000/status \
  -H "Content-Type: application/json" \
  -d '{"status": "confirmed"}'
```

---

### Delete Order

Deletes an order (soft delete).

**Endpoint:** `DELETE /api/v1/orders/{id}`

**Path Parameters:**

| Name | Type | Description |
|------|------|-------------|
| id | uuid | Order ID |

**Response:** `204 No Content`

**Error Responses:**

| Status | Code | Description |
|--------|------|-------------|
| 400 | `MISSING_ID` | No ID provided |
| 404 | `ORDER_NOT_FOUND` | Order does not exist |
| 500 | `INTERNAL_ERROR` | Server error |

**Example:**

```bash
curl -X DELETE http://localhost:8080/api/v1/orders/550e8400-e29b-41d4-a716-446655440000
```

---

## Health Endpoints

### Liveness Probe

Simple check that the server is running.

**Endpoint:** `GET /healthz`

**Response:** `200 OK`

```json
{
  "status": "ok",
  "checks": {},
  "version": "dev"
}
```

**Example:**

```bash
curl http://localhost:8080/healthz
```

---

### Readiness Probe

Checks that the service and its dependencies are ready to handle requests.

**Endpoint:** `GET /readyz`

**Response:** `200 OK` (healthy) or `503 Service Unavailable` (unhealthy)

```json
{
  "status": "ok",
  "checks": {
    "database": "ok"
  },
  "version": "dev"
}
```

**Unhealthy Response (503):**

```json
{
  "status": "unhealthy",
  "checks": {
    "database": "connection refused"
  },
  "version": "dev"
}
```

**Example:**

```bash
curl http://localhost:8080/readyz
```

---

## Error Response Format

All errors return a consistent JSON format:

```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE"
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_REQUEST` | 400 | Malformed request body |
| `MISSING_CUSTOMER_ID` | 400 | customer_id is required |
| `MISSING_ITEMS` | 400 | items array is required |
| `MISSING_ID` | 400 | Order ID is required |
| `MISSING_STATUS` | 400 | status is required |
| `INVALID_CUSTOMER_ID` | 400 | Invalid customer ID format |
| `NO_ITEMS` | 400 | Order must have items |
| `INVALID_TRANSITION` | 400 | Invalid status transition |
| `ORDER_NOT_FOUND` | 404 | Order does not exist |
| `CONCURRENT_MODIFICATION` | 409 | Optimistic lock conflict |
| `INTERNAL_ERROR` | 500 | Internal server error |

---

## Pagination

List endpoints support cursor-based pagination using `limit` and `offset` parameters.

**Response fields:**
- `total` - Total number of records matching the query
- `limit` - Number of records per page (max 100)
- `offset` - Current offset in the result set

**Example pagination flow:**

```bash
# Page 1
curl "http://localhost:8080/api/v1/orders?limit=20&offset=0"

# Page 2
curl "http://localhost:8080/api/v1/orders?limit=20&offset=20"

# Page 3
curl "http://localhost:8080/api/v1/orders?limit=20&offset=40"
```

---

## Versioning

The API uses URL path versioning. All endpoints are prefixed with `/api/v1`.

Future versions will use `/api/v2`, `/api/v3`, etc., allowing for backward-compatible evolution.
