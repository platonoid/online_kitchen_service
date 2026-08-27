package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type Config struct {
	AppPort              string
	DatabaseURL          string
	MigrationsDir        string
	RestaurantServiceURL string
}

type Restaurant struct {
	ID             int     `json:"id"`
	Slug           string  `json:"slug"`
	Name           string  `json:"name"`
	City           string  `json:"city"`
	Cuisine        string  `json:"cuisine"`
	DeliveryEtaMin int     `json:"delivery_eta_min"`
	CommissionRate float64 `json:"commission_rate"`
	IsActive       bool    `json:"is_active"`
}

type MenuItem struct {
	ID           int    `json:"id"`
	RestaurantID int    `json:"restaurant_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	PriceCents   int    `json:"price_cents"`
	IsAvailable  bool   `json:"is_available"`
}

type CustomerInput struct {
	ExternalID  string `json:"external_id,omitempty"`
	DisplayName string `json:"display_name"`
	Phone       string `json:"phone,omitempty"`
}

type OrderItemInput struct {
	ItemID   int `json:"item_id"`
	Quantity int `json:"quantity"`
}

type CreateOrderRequest struct {
	Customer        CustomerInput    `json:"customer"`
	RestaurantID    int              `json:"restaurant_id"`
	DeliveryAddress string           `json:"delivery_address"`
	Notes           string           `json:"notes,omitempty"`
	Items           []OrderItemInput `json:"items"`
}

type OrderItem struct {
	ID             int    `json:"id"`
	OrderID        int    `json:"order_id"`
	ItemID         int    `json:"item_id"`
	Name           string `json:"name"`
	Quantity       int    `json:"quantity"`
	UnitPriceCents int    `json:"unit_price_cents"`
}

type Order struct {
	ID              int           `json:"id"`
	CustomerID      int           `json:"customer_id"`
	RestaurantID    int           `json:"restaurant_id"`
	Status          string        `json:"status"`
	TotalCents      int           `json:"total_cents"`
	DeliveryAddress string        `json:"delivery_address"`
	Notes           string        `json:"notes,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	Items           []OrderItem   `json:"items,omitempty"`
	Customer        CustomerInput `json:"customer,omitempty"`
	Restaurant      *Restaurant   `json:"restaurant,omitempty"`
}

type OrderResponse struct {
	Order                   Order  `json:"order"`
	RestaurantServiceStatus string `json:"restaurant_service_status"`
}

type server struct {
	db                   *sql.DB
	restaurantServiceURL string
}

func main() {
	cfg := loadConfig()

	db, err := connectDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	if err := runMigrations(db, cfg.MigrationsDir); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	srv := &server{db: db, restaurantServiceURL: cfg.RestaurantServiceURL}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/api/v1/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/api/v1/restaurants", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			srv.listRestaurants(w, r)
		default:
			methodNotAllowed(w)
		}
	})
	mux.HandleFunc("/api/v1/restaurants/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/restaurants/")
		if strings.Contains(path, "/menu") {
			srv.getRestaurantMenu(w, r)
			return
		}
		srv.getRestaurantByID(w, r)
	})
	mux.HandleFunc("/api/v1/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			srv.listOrders(w, r)
		case http.MethodPost:
			srv.createOrder(w, r)
		default:
			methodNotAllowed(w)
		}
	})
	mux.HandleFunc("/api/v1/orders/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		srv.getOrderByID(w, r)
	})

	log.Printf("kitchen-api listening on :%s", cfg.AppPort)
	log.Fatal(http.ListenAndServe(":"+cfg.AppPort, mux))
}

func loadConfig() Config {
	migrationsDir := getEnv("MIGRATIONS_DIR", "migrations")
	if _, err := os.Stat(migrationsDir); err != nil {
		if _, err2 := os.Stat("/app/migrations"); err2 == nil {
			migrationsDir = "/app/migrations"
		}
	}

	return Config{
		AppPort:              getEnv("APP_PORT", "8080"),
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/kitchen?sslmode=disable"),
		MigrationsDir:        migrationsDir,
		RestaurantServiceURL: getEnv("RESTAURANT_SERVICE_URL", "http://localhost:8081"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func connectDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	for i := 0; i < 30; i++ {
		if err = db.Ping(); err == nil {
			return db, nil
		}
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("database did not become ready: %w", err)
}

func runMigrations(db *sql.DB, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}
		log.Printf("applied migration: %s", entry.Name())
	}
	return nil
}

func (s *server) listRestaurants(w http.ResponseWriter, _ *http.Request) {
	rows, err := s.db.Query(`
		SELECT id, slug, name, city, cuisine, delivery_eta_min, commission_rate, is_active
		FROM restaurants
		WHERE is_active = TRUE
		ORDER BY name ASC
	`)
	if err != nil {
		internalServerError(w, err)
		return
	}
	defer rows.Close()

	var restaurants []Restaurant
	for rows.Next() {
		var r Restaurant
		if err := rows.Scan(&r.ID, &r.Slug, &r.Name, &r.City, &r.Cuisine, &r.DeliveryEtaMin, &r.CommissionRate, &r.IsActive); err != nil {
			internalServerError(w, err)
			return
		}
		restaurants = append(restaurants, r)
	}
	writeJSON(w, http.StatusOK, restaurants)
}

func (s *server) getRestaurantByID(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r.URL.Path, "/api/v1/restaurants/")
	if err != nil {
		notFound(w)
		return
	}
	var restaurant Restaurant
	err = s.db.QueryRow(`
		SELECT id, slug, name, city, cuisine, delivery_eta_min, commission_rate, is_active
		FROM restaurants
		WHERE id = $1 AND is_active = TRUE`, id).Scan(
		&restaurant.ID, &restaurant.Slug, &restaurant.Name, &restaurant.City, &restaurant.Cuisine,
		&restaurant.DeliveryEtaMin, &restaurant.CommissionRate, &restaurant.IsActive,
	)
	if errors.Is(err, sql.ErrNoRows) {
		notFound(w)
		return
	}
	if err != nil {
		internalServerError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, restaurant)
}

func (s *server) getRestaurantMenu(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/restaurants/")
	if strings.Contains(path, "/menu") {
		path = strings.TrimSuffix(path, "/menu")
	}
	restaurantID, err := strconv.Atoi(strings.TrimSpace(path))
	if err != nil || restaurantID <= 0 {
		notFound(w)
		return
	}
	rows, err := s.db.Query(`
		SELECT id, restaurant_id, name, description, category, price_cents, is_available
		FROM restaurant_items
		WHERE restaurant_id = $1 AND is_available = TRUE
		ORDER BY category ASC, name ASC`, restaurantID)
	if err != nil {
		internalServerError(w, err)
		return
	}
	defer rows.Close()

	var items []MenuItem
	for rows.Next() {
		var item MenuItem
		if err := rows.Scan(&item.ID, &item.RestaurantID, &item.Name, &item.Description, &item.Category, &item.PriceCents, &item.IsAvailable); err != nil {
			internalServerError(w, err)
			return
		}
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"restaurant_id": restaurantID, "items": items})
}

func (s *server) listOrders(w http.ResponseWriter, r *http.Request) {
	customerIDParam := r.URL.Query().Get("customer_id")
	status := r.URL.Query().Get("status")
	query := `
		SELECT o.id, o.customer_id, o.restaurant_id, o.status, o.total_cents, o.delivery_address, o.notes, o.created_at, o.updated_at,
			c.display_name, c.external_id, c.phone,
			r.name, r.slug, r.cuisine, r.city
		FROM orders o
		JOIN customers c ON c.id = o.customer_id
		JOIN restaurants r ON r.id = o.restaurant_id
	`
	args := []any{}
	clauses := []string{}
	if customerIDParam != "" {
		clauses = append(clauses, "o.customer_id = $"+strconv.Itoa(len(args)+1))
		id, err := strconv.Atoi(customerIDParam)
		if err != nil {
			badRequest(w, "customer_id must be int")
			return
		}
		args = append(args, id)
	}
	if status != "" {
		clauses = append(clauses, "o.status = $"+strconv.Itoa(len(args)+1))
		args = append(args, status)
	}
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY o.created_at DESC"
	rows, err := s.db.Query(query, args...)
	if err != nil {
		internalServerError(w, err)
		return
	}
	defer rows.Close()
	var orders []map[string]any
	for rows.Next() {
		var orderID, customerID, restaurantID, totalCents int
		var orderStatus, deliveryAddress, notes, displayName, externalID, phone, restaurantName, restaurantSlug, cuisine, city string
		var createdAt, updatedAt time.Time
		order := map[string]any{}
		if err := rows.Scan(&orderID, &customerID, &restaurantID, &orderStatus, &totalCents, &deliveryAddress, &notes, &createdAt, &updatedAt,
			&displayName, &externalID, &phone, &restaurantName, &restaurantSlug, &cuisine, &city); err != nil {
			internalServerError(w, err)
			return
		}
		order["id"] = orderID
		order["customer_id"] = customerID
		order["restaurant_id"] = restaurantID
		order["status"] = orderStatus
		order["total_cents"] = totalCents
		order["delivery_address"] = deliveryAddress
		order["notes"] = notes
		order["created_at"] = createdAt
		order["updated_at"] = updatedAt
		order["customer"] = map[string]any{"display_name": displayName, "external_id": externalID, "phone": phone}
		order["restaurant"] = map[string]any{"name": restaurantName, "slug": restaurantSlug, "cuisine": cuisine, "city": city}
		orders = append(orders, order)
	}
	writeJSON(w, http.StatusOK, orders)
}

func (s *server) getOrderByID(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r.URL.Path, "/api/v1/orders/")
	if err != nil {
		notFound(w)
		return
	}
	order := Order{}
	if err := s.db.QueryRow(`
		SELECT o.id, o.customer_id, o.restaurant_id, o.status, o.total_cents, o.delivery_address, o.notes, o.created_at, o.updated_at
		FROM orders o WHERE o.id = $1`, id).Scan(
		&order.ID, &order.CustomerID, &order.RestaurantID, &order.Status, &order.TotalCents,
		&order.DeliveryAddress, &order.Notes, &order.CreatedAt, &order.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w)
			return
		}
		internalServerError(w, err)
		return
	}
	customerRow := CustomerInput{}
	if err := s.db.QueryRow(`SELECT external_id, display_name, phone FROM customers WHERE id = $1`, order.CustomerID).Scan(&customerRow.ExternalID, &customerRow.DisplayName, &customerRow.Phone); err != nil {
		internalServerError(w, err)
		return
	}
	order.Customer = customerRow
	restaurantRow := Restaurant{}
	if err := s.db.QueryRow(`SELECT id, slug, name, city, cuisine, delivery_eta_min, commission_rate, is_active FROM restaurants WHERE id = $1`, order.RestaurantID).Scan(
		&restaurantRow.ID, &restaurantRow.Slug, &restaurantRow.Name, &restaurantRow.City, &restaurantRow.Cuisine,
		&restaurantRow.DeliveryEtaMin, &restaurantRow.CommissionRate, &restaurantRow.IsActive,
	); err != nil {
		internalServerError(w, err)
		return
	}
	order.Restaurant = &restaurantRow
	rows, err := s.db.Query(`
		SELECT id, order_id, item_id, name_snapshot, quantity, unit_price_cents
		FROM order_items WHERE order_id = $1 ORDER BY id ASC`, order.ID)
	if err != nil {
		internalServerError(w, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var line OrderItem
		if err := rows.Scan(&line.ID, &line.OrderID, &line.ItemID, &line.Name, &line.Quantity, &line.UnitPriceCents); err != nil {
			internalServerError(w, err)
			return
		}
		order.Items = append(order.Items, line)
	}
	writeJSON(w, http.StatusOK, order)
}

func (s *server) createOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid JSON payload")
		return
	}
	if req.RestaurantID <= 0 || len(req.Items) == 0 || strings.TrimSpace(req.DeliveryAddress) == "" {
		badRequest(w, "restaurant_id, delivery_address and at least one item are required")
		return
	}
	if strings.TrimSpace(req.Customer.DisplayName) == "" {
		badRequest(w, "customer.display_name is required")
		return
	}

	tx, err := s.db.Begin()
	if err != nil {
		internalServerError(w, err)
		return
	}
	defer tx.Rollback()

	customerID, err := upsertCustomer(tx, req.Customer)
	if err != nil {
		internalServerError(w, err)
		return
	}

	var restaurant Restaurant
	err = tx.QueryRow(`
		SELECT id, slug, name, city, cuisine, delivery_eta_min, commission_rate, is_active
		FROM restaurants WHERE id = $1 AND is_active = TRUE`, req.RestaurantID).Scan(
		&restaurant.ID, &restaurant.Slug, &restaurant.Name, &restaurant.City, &restaurant.Cuisine,
		&restaurant.DeliveryEtaMin, &restaurant.CommissionRate, &restaurant.IsActive,
	)
	if errors.Is(err, sql.ErrNoRows) {
		badRequest(w, "restaurant not found or inactive")
		return
	}
	if err != nil {
		internalServerError(w, err)
		return
	}

	totalCents := 0
	orderItems := make([]OrderItem, 0, len(req.Items))
	for _, itemReq := range req.Items {
		if itemReq.ItemID <= 0 || itemReq.Quantity <= 0 {
			badRequest(w, "each item requires item_id and quantity > 0")
			return
		}
		var item MenuItem
		err = tx.QueryRow(`
			SELECT id, restaurant_id, name, description, category, price_cents, is_available
			FROM restaurant_items WHERE id = $1 AND restaurant_id = $2 AND is_available = TRUE`, itemReq.ItemID, req.RestaurantID).Scan(
			&item.ID, &item.RestaurantID, &item.Name, &item.Description, &item.Category, &item.PriceCents, &item.IsAvailable,
		)
		if errors.Is(err, sql.ErrNoRows) {
			badRequest(w, fmt.Sprintf("item %d is not available for this restaurant", itemReq.ItemID))
			return
		}
		if err != nil {
			internalServerError(w, err)
			return
		}
		totalCents += item.PriceCents * itemReq.Quantity
		orderItems = append(orderItems, OrderItem{
			ItemID:         item.ID,
			Name:           item.Name,
			Quantity:       itemReq.Quantity,
			UnitPriceCents: item.PriceCents,
		})
	}

	var orderID int
	err = tx.QueryRow(`
		INSERT INTO orders (customer_id, restaurant_id, status, total_cents, delivery_address, notes)
		VALUES ($1, $2, 'pending', $3, $4, $5)
		RETURNING id`, customerID, req.RestaurantID, totalCents, req.DeliveryAddress, req.Notes).Scan(&orderID)
	if err != nil {
		internalServerError(w, err)
		return
	}
	for _, item := range orderItems {
		if _, err = tx.Exec(`
			INSERT INTO order_items (order_id, item_id, name_snapshot, quantity, unit_price_cents)
			VALUES ($1, $2, $3, $4, $5)`, orderID, item.ItemID, item.Name, item.Quantity, item.UnitPriceCents); err != nil {
			internalServerError(w, err)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		internalServerError(w, err)
		return
	}

	order := Order{
		ID:              orderID,
		CustomerID:      customerID,
		RestaurantID:    req.RestaurantID,
		Status:          "pending",
		TotalCents:      totalCents,
		DeliveryAddress: req.DeliveryAddress,
		Notes:           req.Notes,
		Customer:        req.Customer,
		Restaurant:      &restaurant,
		Items:           orderItems,
	}
	restaurantStatus, err := s.notifyRestaurant(order)
	if err != nil {
		restaurantStatus = "failed"
		log.Printf("restaurant notification failed for order %d: %v", orderID, err)
	}
	writeJSON(w, http.StatusCreated, OrderResponse{Order: order, RestaurantServiceStatus: restaurantStatus})
}

func upsertCustomer(tx *sql.Tx, customer CustomerInput) (int, error) {
	customer.DisplayName = strings.TrimSpace(customer.DisplayName)
	customer.Phone = strings.TrimSpace(customer.Phone)
	if customer.ExternalID == "" {
		var customerID int
		if err := tx.QueryRow(`
			INSERT INTO customers (display_name, phone)
			VALUES ($1, $2)
			RETURNING id`, customer.DisplayName, customer.Phone).Scan(&customerID); err != nil {
			return 0, err
		}
		return customerID, nil
	}
	var customerID int
	err := tx.QueryRow(`
		SELECT id FROM customers WHERE external_id = $1`, customer.ExternalID).Scan(&customerID)
	if errors.Is(err, sql.ErrNoRows) {
		if err = tx.QueryRow(`
			INSERT INTO customers (external_id, display_name, phone)
			VALUES ($1, $2, $3)
			RETURNING id`, customer.ExternalID, customer.DisplayName, customer.Phone).Scan(&customerID); err != nil {
			return 0, err
		}
		return customerID, nil
	}
	if err != nil {
		return 0, err
	}
	if _, err = tx.Exec(`UPDATE customers SET display_name = $1, phone = $2 WHERE id = $3`, customer.DisplayName, customer.Phone, customerID); err != nil {
		return 0, err
	}
	return customerID, nil
}

func (s *server) notifyRestaurant(order Order) (string, error) {
	payload := map[string]any{
		"order_id":       order.ID,
		"restaurant_id":  order.RestaurantID,
		"status":         order.Status,
		"customer_name":  order.Customer.DisplayName,
		"customer_phone": order.Customer.Phone,
		"customer": map[string]any{
			"display_name": order.Customer.DisplayName,
			"external_id":  order.Customer.ExternalID,
			"phone":        order.Customer.Phone,
		},
		"delivery_address": order.DeliveryAddress,
		"notes":            order.Notes,
		"total_cents":      order.TotalCents,
		"items": func() []map[string]any {
			out := make([]map[string]any, 0, len(order.Items))
			for _, item := range order.Items {
				out = append(out, map[string]any{
					"item_id":          item.ItemID,
					"name":             item.Name,
					"quantity":         item.Quantity,
					"unit_price_cents": item.UnitPriceCents,
				})
			}
			return out
		}(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	url := strings.TrimRight(s.restaurantServiceURL, "/") + "/api/v1/integrations/order-received"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		var errBody map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		return "", fmt.Errorf("restaurant service returned %d: %v", resp.StatusCode, errBody)
	}
	return "accepted", nil
}

func parsePathID(path string, prefix string) (int, error) {
	trimmed := strings.TrimPrefix(path, prefix)
	trimmed = strings.Trim(trimmed, "/")
	if trimmed == "" {
		return 0, errors.New("missing id")
	}
	id, err := strconv.Atoi(trimmed)
	if err != nil || id <= 0 {
		return 0, err
	}
	return id, nil
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write JSON: %v", err)
	}
}

func methodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func badRequest(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": message})
}

func notFound(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "resource not found"})
}

func internalServerError(w http.ResponseWriter, err error) {
	log.Printf("internal server error: %v", err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}
