package http

import (
	"net/http"
	"strconv"
	"strings"

	"kitchen/internal/dto"
	"kitchen/internal/services"
)

func (h *Handler) handleAllDishes(w http.ResponseWriter, r *http.Request, parts []string) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	if len(parts) > 0 {
		if len(parts) != 1 {
			notFound(w)
			return
		}
		dishID, err := strconv.Atoi(parts[0])
		if err != nil {
			notFound(w)
			return
		}
		dish, serviceErr := h.catalog.GetDish(dishID)
		if serviceErr != nil {
			writeServiceError(w, serviceErr)
			return
		}
		writeJSON(w, http.StatusOK, dto.DishFromDomain(dish))
		return
	}
	shopID, _ := strconv.Atoi(r.URL.Query().Get("shop_id"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	dishes, total, err := h.catalog.ListDishes(shopID, services.DishFilter{Query: r.URL.Query().Get("q"), Category: r.URL.Query().Get("type"), Sort: r.URL.Query().Get("sort"), Page: page, PageSize: pageSize})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.DishListResponse{
		Items: dto.DishesFromDomain(dishes), Total: total,
		Page: normalizedPage(page), PageSize: normalizedPageSize(pageSize),
	})
}

func (h *Handler) handleShop(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		writeJSON(w, http.StatusOK, dto.ShopsFromDomain(h.catalog.ListShops()))
		return
	}
	shopID, err := strconv.Atoi(parts[1])
	if err != nil {
		notFound(w)
		return
	}
	if len(parts) == 2 {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		shop, serviceErr := h.catalog.GetShop(shopID)
		if serviceErr != nil {
			writeServiceError(w, serviceErr)
			return
		}
		writeJSON(w, http.StatusOK, dto.ShopFromDomain(shop))
		return
	}
	switch parts[2] {
	case "dishes", "menu":
		h.handleDishes(w, r, shopID)
	default:
		notFound(w)
	}
}

func (h *Handler) handleDishes(w http.ResponseWriter, r *http.Request, shopID int) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize == 0 {
		pageSize, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	}
	if page == 0 {
		page, _ = strconv.Atoi(r.URL.Query().Get("page_number"))
	}
	query := r.URL.Query().Get("q")
	if query == "" {
		query = r.URL.Query().Get("search")
	}
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = r.URL.Query().Get("sort_by")
	}
	if r.URL.Query().Get("order") == "desc" && !strings.HasSuffix(sortBy, "_desc") {
		sortBy += "_desc"
	}
	filter := services.DishFilter{Query: query, Category: r.URL.Query().Get("category"), Sort: sortBy, Page: page, PageSize: pageSize}
	dishes, total, err := h.catalog.ListDishes(shopID, filter)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	payload := dto.DishListResponse{Items: dto.DishesFromDomain(dishes), Total: total, Page: normalizedPage(page), PageSize: normalizedPageSize(pageSize)}
	if strings.Contains(r.URL.Path, "/menu") {
		writeJSON(w, http.StatusOK, map[string]any{"shop_id": shopID, "items": payload.Items, "total": payload.Total, "page": payload.Page, "page_size": payload.PageSize})
		return
	}
	writeJSON(w, http.StatusOK, payload)
}
