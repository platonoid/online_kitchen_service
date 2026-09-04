package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"kitchen/internal/repository"
	"kitchen/internal/services"
)

func splitPath(path string) []string {
	raw := strings.Split(strings.Trim(path, "/"), "/")
	if len(raw) == 1 && raw[0] == "" {
		return nil
	}
	return raw
}

func normalizedPage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func normalizedPageSize(size int) int {
	if size < 1 {
		return 20
	}
	return size
}

func decodeJSON(r *http.Request, value any) error {
	return json.NewDecoder(r.Body).Decode(value)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeServiceError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, repository.ErrNotFound) {
		status = http.StatusNotFound
	} else if errors.Is(err, services.ErrInvalidQuantity) || errors.Is(err, services.ErrIngredientNotRemovable) {
		status = http.StatusBadRequest
	} else if errors.Is(err, services.ErrDishUnavailable) {
		status = http.StatusConflict
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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
