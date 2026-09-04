-- name: ListActiveShops :many
SELECT id, name, city, cuisine, delivery_time_min, is_active
FROM shops
WHERE is_active = TRUE
ORDER BY id;

-- name: ListAvailableDishes :many
SELECT id, shop_id, name, description, category, price_cents, preparation_time_min, is_available
FROM dishes
WHERE is_available = TRUE
  AND ($1::integer = 0 OR shop_id = $1)
ORDER BY id;
