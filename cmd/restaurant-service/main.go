package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

type RestaurantInfo struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	City           string `json:"city"`
	Cuisine        string `json:"cuisine"`
	DeliveryEtaMin int    `json:"delivery_eta_min"`
}

type RestaurantOrder struct {
	OrderID         int              `json:"order_id"`
	RestaurantID    int              `json:"restaurant_id"`
	Status          string           `json:"status"`
	CustomerName    string           `json:"customer_name"`
	CustomerPhone   string           `json:"customer_phone,omitempty"`
	DeliveryAddress string           `json:"delivery_address"`
	Notes           string           `json:"notes,omitempty"`
	TotalCents      int              `json:"total_cents"`
	Items           []map[string]any `json:"items"`
}

var (
	mu     sync.Mutex
	orders = make([]RestaurantOrder, 0)
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/api/v1/restaurants", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		restaurants := []RestaurantInfo{
			{ID: 1, Name: "Chai Corner", City: "Moscow", Cuisine: "Indian", DeliveryEtaMin: 25},
			{ID: 2, Name: "PastaLab", City: "Moscow", Cuisine: "Italian", DeliveryEtaMin: 30},
			{ID: 3, Name: "Sushi Metro", City: "Moscow", Cuisine: "Japanese", DeliveryEtaMin: 22},
		}
		writeJSON(w, http.StatusOK, restaurants)
	})
	mux.HandleFunc("/api/v1/integrations/order-received", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			methodNotAllowed(w)
			return
		}
		var payload RestaurantOrder
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			badRequest(w, "invalid JSON payload")
			return
		}
		if payload.OrderID == 0 || payload.RestaurantID == 0 {
			badRequest(w, "order_id and restaurant_id are required")
			return
		}
		mu.Lock()
		orders = append(orders, payload)
		mu.Unlock()
		writeJSON(w, http.StatusAccepted, map[string]any{"status": "accepted", "order_id": payload.OrderID})
	})
	mux.HandleFunc("/api/v1/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		mu.Lock()
		defer mu.Unlock()
		writeJSON(w, http.StatusOK, orders)
	})
	mux.HandleFunc("/api/v1/restaurants/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		parts := strings.TrimPrefix(r.URL.Path, "/api/v1/restaurants/")
		parts = strings.Trim(parts, "/")
		if parts == "" {
			restaurants := []RestaurantInfo{
				{ID: 1, Name: "Chai Corner", City: "Moscow", Cuisine: "Indian", DeliveryEtaMin: 25},
				{ID: 2, Name: "PastaLab", City: "Moscow", Cuisine: "Italian", DeliveryEtaMin: 30},
				{ID: 3, Name: "Sushi Metro", City: "Moscow", Cuisine: "Japanese", DeliveryEtaMin: 22},
			}
			writeJSON(w, http.StatusOK, restaurants)
			return
		}
		restaurantID, err := strconv.Atoi(strings.Split(parts, "/")[0])
		if err != nil {
			notFound(w)
			return
		}
		for _, item := range []RestaurantInfo{{ID: 1, Name: "Chai Corner", City: "Moscow", Cuisine: "Indian", DeliveryEtaMin: 25}, {ID: 2, Name: "PastaLab", City: "Moscow", Cuisine: "Italian", DeliveryEtaMin: 30}, {ID: 3, Name: "Sushi Metro", City: "Moscow", Cuisine: "Japanese", DeliveryEtaMin: 22}} {
			if restaurantID == item.ID {
				writeJSON(w, http.StatusOK, item)
				return
			}
		}
		notFound(w)
	})
	port := getEnv("APP_PORT", "8081")
	log.Printf("restaurant-service listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func badRequest(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": message})
}

func methodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func notFound(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "resource not found"})
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
