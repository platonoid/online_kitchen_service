package repository

import (
	"database/sql"
	"fmt"
	"sync/atomic"

	_ "github.com/jackc/pgx/v5/stdlib"
	"kitchen/internal/domain"
)

type PostgresRepository struct {
	db         *sql.DB
	nextItemID atomic.Int64
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func OpenPostgresRepository(dsn string) (*PostgresRepository, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return NewPostgresRepository(db), nil
}

func (r *PostgresRepository) ListShops() []domain.Shop {
	rows, err := r.db.Query(`SELECT id, name, city, cuisine, delivery_time_min, is_active FROM shops WHERE is_active = TRUE ORDER BY id`)
	if err != nil {
		return []domain.Shop{}
	}
	defer rows.Close()
	shops := make([]domain.Shop, 0)
	for rows.Next() {
		var shop domain.Shop
		if rows.Scan(&shop.ID, &shop.Name, &shop.City, &shop.Cuisine, &shop.DeliveryTimeMin, &shop.IsActive) == nil {
			shops = append(shops, shop)
		}
	}
	return shops
}

func (r *PostgresRepository) GetShop(id int) (domain.Shop, error) {
	var shop domain.Shop
	err := r.db.QueryRow(`SELECT id, name, city, cuisine, delivery_time_min, is_active FROM shops WHERE id = $1 AND is_active = TRUE`, id).
		Scan(&shop.ID, &shop.Name, &shop.City, &shop.Cuisine, &shop.DeliveryTimeMin, &shop.IsActive)
	if err == sql.ErrNoRows {
		return domain.Shop{}, ErrNotFound
	}
	if err != nil {
		return domain.Shop{}, fmt.Errorf("get shop: %w", err)
	}
	return shop, nil
}

func (r *PostgresRepository) ListDishes(shopID int) []domain.Dish {
	query := `SELECT id, shop_id, name, COALESCE(description, ''), category, price_cents, preparation_time_min, is_available
		FROM dishes WHERE is_available = TRUE`
	args := []any{}
	if shopID != 0 {
		query += ` AND shop_id = $1`
		args = append(args, shopID)
	}
	query += ` ORDER BY id`
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return []domain.Dish{}
	}
	defer rows.Close()
	dishes := make([]domain.Dish, 0)
	for rows.Next() {
		var dish domain.Dish
		if rows.Scan(&dish.ID, &dish.ShopID, &dish.Name, &dish.Description, &dish.Category, &dish.PriceCents, &dish.PreparationTimeMin, &dish.IsAvailable) == nil {
			dish.Ingredients = r.ingredients(dish.ID)
			dishes = append(dishes, dish)
		}
	}
	return dishes
}

func (r *PostgresRepository) GetDish(id int) (domain.Dish, error) {
	var dish domain.Dish
	err := r.db.QueryRow(`SELECT id, shop_id, name, COALESCE(description, ''), category, price_cents, preparation_time_min, is_available
		FROM dishes WHERE id = $1 AND is_available = TRUE`, id).
		Scan(&dish.ID, &dish.ShopID, &dish.Name, &dish.Description, &dish.Category, &dish.PriceCents, &dish.PreparationTimeMin, &dish.IsAvailable)
	if err == sql.ErrNoRows {
		return domain.Dish{}, ErrNotFound
	}
	if err != nil {
		return domain.Dish{}, fmt.Errorf("get dish: %w", err)
	}
	dish.Ingredients = r.ingredients(dish.ID)
	return dish, nil
}

func (r *PostgresRepository) ingredients(dishID int) []domain.Ingredient {
	rows, err := r.db.Query(`SELECT i.id, i.name, di.removable FROM ingredients i JOIN dish_ingredients di ON di.ingredient_id = i.id WHERE di.dish_id = $1 ORDER BY i.id`, dishID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	ingredients := make([]domain.Ingredient, 0)
	for rows.Next() {
		var ingredient domain.Ingredient
		ingredient.DishID = dishID
		if rows.Scan(&ingredient.ID, &ingredient.Name, &ingredient.Removable) == nil {
			ingredients = append(ingredients, ingredient)
		}
	}
	return ingredients
}

func (r *PostgresRepository) CreateOrder(order domain.Order) (domain.Order, error) {
	err := r.db.QueryRow(`INSERT INTO orders (shop_id, status, total_cents, estimated_time_min) VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`, order.ShopID, order.Status, order.TotalCents, order.EstimatedTimeMin).
		Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err == sql.ErrNoRows {
		return domain.Order{}, ErrNotFound
	}
	if err != nil {
		return domain.Order{}, fmt.Errorf("create order: %w", err)
	}
	order.Items = []domain.OrderItem{}
	return order, nil
}

func (r *PostgresRepository) GetOrder(id int) (domain.Order, error) {
	var order domain.Order
	err := r.db.QueryRow(`SELECT id, shop_id, status, total_cents, estimated_time_min, created_at, updated_at FROM orders WHERE id = $1`, id).
		Scan(&order.ID, &order.ShopID, &order.Status, &order.TotalCents, &order.EstimatedTimeMin, &order.CreatedAt, &order.UpdatedAt)
	if err == sql.ErrNoRows {
		return domain.Order{}, ErrNotFound
	}
	if err != nil {
		return domain.Order{}, fmt.Errorf("get order: %w", err)
	}
	order.Items = r.orderItems(id)
	return order, nil
}

func (r *PostgresRepository) orderItems(orderID int) []domain.OrderItem {
	rows, err := r.db.Query(`SELECT id, dish_id, dish_name_snapshot, quantity, unit_price_cents, status, COALESCE(rejection_reason, '') FROM order_items WHERE order_id = $1 ORDER BY id`, orderID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	items := make([]domain.OrderItem, 0)
	for rows.Next() {
		var item domain.OrderItem
		item.OrderID = orderID
		if rows.Scan(&item.ID, &item.DishID, &item.DishName, &item.Quantity, &item.UnitPriceCents, &item.Status, &item.RejectionReason) == nil {
			item.ExcludedIngredientIDs = r.excludedIngredients(item.ID)
			items = append(items, item)
		}
	}
	return items
}

func (r *PostgresRepository) excludedIngredients(itemID int) []int {
	rows, err := r.db.Query(`SELECT ingredient_id FROM order_item_excluded_ingredients WHERE order_item_id = $1 ORDER BY ingredient_id`, itemID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	ids := make([]int, 0)
	for rows.Next() {
		var id int
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

func (r *PostgresRepository) SaveOrder(order domain.Order) error {
	result, err := r.db.Exec(`UPDATE orders SET status = $1, total_cents = $2, estimated_time_min = $3, updated_at = NOW() WHERE id = $4`, order.Status, order.TotalCents, order.EstimatedTimeMin, order.ID)
	if err != nil {
		return fmt.Errorf("save order: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		return ErrNotFound
	}
	if _, err := r.db.Exec(`DELETE FROM order_items WHERE order_id = $1`, order.ID); err != nil {
		return fmt.Errorf("replace order items: %w", err)
	}
	for _, item := range order.Items {
		if _, err := r.db.Exec(`INSERT INTO order_items (id, order_id, dish_id, dish_name_snapshot, unit_price_cents, quantity, status, rejection_reason) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`, item.ID, order.ID, item.DishID, item.DishName, item.UnitPriceCents, item.Quantity, item.Status, item.RejectionReason); err != nil {
			return fmt.Errorf("save order item: %w", err)
		}
		for _, ingredientID := range item.ExcludedIngredientIDs {
			if _, err := r.db.Exec(`INSERT INTO order_item_excluded_ingredients (order_item_id, ingredient_id) VALUES ($1, $2)`, item.ID, ingredientID); err != nil {
				return fmt.Errorf("save excluded ingredient: %w", err)
			}
		}
	}
	return nil
}

func (r *PostgresRepository) NextItemID() int {
	if id := r.nextItemID.Add(1); id > 1 {
		return int(id)
	}
	return 1
}
