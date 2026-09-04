package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kitchen/internal/kafka"
	"kitchen/internal/repository"
	"kitchen/internal/services"
)

func testHandler() http.Handler {
	repo := repository.NewInMemoryRepository()
	catalog := services.NewCatalogService(repo)
	orders := services.NewOrderService(repo, kafka.NewLoggingPublisher(nil))
	return NewHandler(catalog, orders)
}

func TestHandlerCreatesEmptyOrderAndChangesItems(t *testing.T) {
	handler := testHandler()
	create := httptest.NewRequest(http.MethodPost, "/api/v1/orders", strings.NewReader(`{"shop_id":1}`))
	create.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, create)
	if response.Code != http.StatusCreated || !strings.Contains(response.Body.String(), `"total_cents":0`) {
		t.Fatalf("create response: %d %s", response.Code, response.Body.String())
	}
	add := httptest.NewRequest(http.MethodPost, "/api/v1/orders/1/items", strings.NewReader(`{"dish_id":1,"quantity":2,"excluded_ingredient_ids":[1]}`))
	add.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, add)
	if response.Code != http.StatusCreated || !strings.Contains(response.Body.String(), `"total_cents":500`) {
		t.Fatalf("add response: %d %s", response.Code, response.Body.String())
	}
	invalid := httptest.NewRequest(http.MethodPost, "/api/v1/orders/1/items", strings.NewReader(`{"dish_id":2,"quantity":1,"excluded_ingredient_ids":[4]}`))
	invalid.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, invalid)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("invalid ingredient response: %d", response.Code)
	}
}

func TestHandlerDishPaginationAndEstimate(t *testing.T) {
	handler := testHandler()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/shops/1/dishes?category=main&page_size=1&sort=price_desc", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"total":2`) {
		t.Fatalf("dish response: %d %s", response.Code, response.Body.String())
	}
	create := httptest.NewRequest(http.MethodPost, "/api/v1/orders", strings.NewReader(`{"shop_id":1}`))
	create.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, create)
	if response.Code != http.StatusCreated {
		t.Fatalf("order response: %d", response.Code)
	}
	estimate := httptest.NewRequest(http.MethodGet, "/api/v1/orders/1/estimate", nil)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, estimate)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"estimated_time_min":25`) {
		t.Fatalf("estimate response: %d %s", response.Code, response.Body.String())
	}
}
