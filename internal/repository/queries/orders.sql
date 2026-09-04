-- name: CreateOrder :one
INSERT INTO orders (shop_id, status, total_cents, estimated_time_min)
VALUES ($1, $2, $3, $4)
RETURNING id, shop_id, status, total_cents, estimated_time_min, created_at, updated_at;

-- name: GetOrder :one
SELECT id, shop_id, status, total_cents, estimated_time_min, created_at, updated_at
FROM orders
WHERE id = $1;
