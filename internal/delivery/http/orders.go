package http

import (
	"net/http"
	"strconv"

	"kitchen/internal/dto"
	"kitchen/internal/services"
)

func (h *Handler) handleOrder(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 1 {
		if r.Method != http.MethodPost {
			methodNotAllowed(w)
			return
		}
		var request dto.CreateOrderRequest
		if err := decodeJSON(r, &request); err != nil || request.ShopID < 1 {
			badRequest(w, "shop_id is required")
			return
		}
		order, err := h.orders.CreateOrder(r.Context(), request.ShopID)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, dto.OrderFromDomain(order))
		return
	}
	orderID, err := strconv.Atoi(parts[1])
	if err != nil {
		notFound(w)
		return
	}
	if len(parts) == 2 {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		order, serviceErr := h.orders.GetOrder(r.Context(), orderID)
		if serviceErr != nil {
			writeServiceError(w, serviceErr)
			return
		}
		writeJSON(w, http.StatusOK, dto.OrderFromDomain(order))
		return
	}
	switch parts[2] {
	case "items":
		h.handleItems(w, r, orderID, parts[3:])
	case "estimate", "estimate-time":
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		estimate, serviceErr := h.orders.EstimateTime(r.Context(), orderID)
		if serviceErr != nil {
			writeServiceError(w, serviceErr)
			return
		}
		writeJSON(w, http.StatusOK, dto.EstimateResponse{EstimatedTimeMin: estimate})
	default:
		notFound(w)
	}
}

func (h *Handler) handleItems(w http.ResponseWriter, r *http.Request, orderID int, parts []string) {
	if len(parts) == 0 {
		if r.Method != http.MethodPost {
			methodNotAllowed(w)
			return
		}
		var request dto.AddOrderItemRequest
		if err := decodeJSON(r, &request); err != nil {
			badRequest(w, "invalid JSON payload")
			return
		}
		if request.DishID == 0 {
			request.DishID = request.ItemID
		}
		if request.DishID < 1 {
			badRequest(w, "dish_id is required")
			return
		}
		order, serviceErr := h.orders.AddItem(r.Context(), orderID, services.AddItemInput{DishID: request.DishID, Quantity: request.Quantity, ExcludedIngredientIDs: request.ExcludedIngredientIDs})
		if serviceErr != nil {
			writeServiceError(w, serviceErr)
			return
		}
		writeJSON(w, http.StatusCreated, dto.OrderFromDomain(order))
		return
	}
	itemID, err := strconv.Atoi(parts[0])
	if err != nil {
		notFound(w)
		return
	}
	switch r.Method {
	case http.MethodPatch, http.MethodPut:
		var request dto.UpdateOrderItemRequest
		if err := decodeJSON(r, &request); err != nil {
			badRequest(w, "invalid JSON payload")
			return
		}
		order, serviceErr := h.orders.UpdateItem(r.Context(), orderID, itemID, services.UpdateItemInput{
			Quantity: request.Quantity, ExcludedIngredientIDs: request.ExcludedIngredientIDs,
		})
		if serviceErr != nil {
			writeServiceError(w, serviceErr)
			return
		}
		writeJSON(w, http.StatusOK, dto.OrderFromDomain(order))
	case http.MethodDelete:
		order, serviceErr := h.orders.DeleteItem(r.Context(), orderID, itemID)
		if serviceErr != nil {
			writeServiceError(w, serviceErr)
			return
		}
		writeJSON(w, http.StatusOK, dto.OrderFromDomain(order))
	default:
		methodNotAllowed(w)
	}
}
