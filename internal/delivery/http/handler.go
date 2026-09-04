package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"kitchen/internal/metrics"
	"kitchen/internal/services"
)

type Handler struct {
	catalog *services.CatalogService
	orders  *services.OrderService
	router  http.Handler
}

func NewHandler(catalog *services.CatalogService, orders *services.OrderService) *Handler {
	handler := &Handler{catalog: catalog, orders: orders}
	handler.router = handler.newRouter()
	return handler
}

func NewHTTPHandler(catalog *services.CatalogService, orders *services.OrderService) http.Handler {
	return NewHandler(catalog, orders)
}

func NewRouter(catalog *services.CatalogService, orders *services.OrderService) http.Handler {
	return NewHandler(catalog, orders)
}

func (h *Handler) newRouter() http.Handler {
	router := chi.NewRouter()
	router.Use(metrics.Middleware)
	router.Handle("/api/v1", http.HandlerFunc(h.handleAPI))
	router.Handle("/api/v1/*", http.HandlerFunc(h.handleAPI))
	router.Handle("/healthz", http.HandlerFunc(health))
	router.Handle("/health", http.HandlerFunc(health))
	router.Handle("/metrics", promhttp.Handler())
	router.Handle("/api/v1/metrics", promhttp.Handler())
	return router
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *Handler) handleAPI(w http.ResponseWriter, r *http.Request) {
	path := trimAPIPath(r.URL.Path)
	if path == "healthz" || path == "health" {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}
	parts := splitPath(path)
	if len(parts) == 0 {
		notFound(w)
		return
	}
	switch parts[0] {
	case "shops":
		h.handleShop(w, r, parts)
	case "dishes":
		h.handleAllDishes(w, r, parts[1:])
	case "orders":
		h.handleOrder(w, r, parts)
	default:
		notFound(w)
	}
}

func trimAPIPath(path string) string {
	if len(path) >= len("/api/v1") && path[:len("/api/v1")] == "/api/v1" {
		path = path[len("/api/v1"):]
	}
	for len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	return path
}
