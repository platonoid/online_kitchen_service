package dto

import (
	"time"

	"kitchen/internal/domain"
)

type ShopResponse struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	City            string `json:"city"`
	Cuisine         string `json:"cuisine"`
	DeliveryTimeMin int    `json:"delivery_time_min"`
	IsActive        bool   `json:"is_active"`
}

type IngredientResponse struct {
	ID        int    `json:"id"`
	DishID    int    `json:"dish_id"`
	Name      string `json:"name"`
	Removable bool   `json:"removable"`
}

type DishResponse struct {
	ID                 int                  `json:"id"`
	ShopID             int                  `json:"shop_id"`
	Name               string               `json:"name"`
	Description        string               `json:"description,omitempty"`
	Category           string               `json:"category"`
	PriceCents         int                  `json:"price_cents"`
	PreparationTimeMin int                  `json:"preparation_time_min"`
	IsAvailable        bool                 `json:"is_available"`
	Ingredients        []IngredientResponse `json:"ingredients,omitempty"`
}

type DishListResponse struct {
	Items    []DishResponse `json:"items"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

type OrderItemResponse struct {
	ID                    int    `json:"id"`
	OrderID               int    `json:"order_id"`
	DishID                int    `json:"dish_id"`
	DishName              string `json:"dish_name"`
	Quantity              int    `json:"quantity"`
	UnitPriceCents        int    `json:"unit_price_cents"`
	Status                string `json:"status"`
	RejectionReason       string `json:"rejection_reason,omitempty"`
	ExcludedIngredientIDs []int  `json:"excluded_ingredient_ids,omitempty"`
}

type OrderResponse struct {
	ID               int                 `json:"id"`
	ShopID           int                 `json:"shop_id"`
	Status           string              `json:"status"`
	Items            []OrderItemResponse `json:"items"`
	TotalCents       int                 `json:"total_cents"`
	EstimatedTimeMin int                 `json:"estimated_time_min"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

type EstimateResponse struct {
	EstimatedTimeMin int `json:"estimated_time_min"`
}

func ShopFromDomain(shop domain.Shop) ShopResponse {
	return ShopResponse{
		ID: shop.ID, Name: shop.Name, City: shop.City, Cuisine: shop.Cuisine,
		DeliveryTimeMin: shop.DeliveryTimeMin, IsActive: shop.IsActive,
	}
}

func ShopsFromDomain(shops []domain.Shop) []ShopResponse {
	result := make([]ShopResponse, 0, len(shops))
	for _, shop := range shops {
		result = append(result, ShopFromDomain(shop))
	}
	return result
}

func DishFromDomain(dish domain.Dish) DishResponse {
	ingredients := make([]IngredientResponse, 0, len(dish.Ingredients))
	for _, ingredient := range dish.Ingredients {
		ingredients = append(ingredients, IngredientResponse{
			ID: ingredient.ID, DishID: ingredient.DishID, Name: ingredient.Name, Removable: ingredient.Removable,
		})
	}
	return DishResponse{
		ID: dish.ID, ShopID: dish.ShopID, Name: dish.Name, Description: dish.Description,
		Category: dish.Category, PriceCents: dish.PriceCents, PreparationTimeMin: dish.PreparationTimeMin,
		IsAvailable: dish.IsAvailable, Ingredients: ingredients,
	}
}

func DishesFromDomain(dishes []domain.Dish) []DishResponse {
	result := make([]DishResponse, 0, len(dishes))
	for _, dish := range dishes {
		result = append(result, DishFromDomain(dish))
	}
	return result
}

func OrderFromDomain(order domain.Order) OrderResponse {
	items := make([]OrderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, OrderItemResponse{
			ID: item.ID, OrderID: item.OrderID, DishID: item.DishID, DishName: item.DishName,
			Quantity: item.Quantity, UnitPriceCents: item.UnitPriceCents, Status: item.Status,
			RejectionReason:       item.RejectionReason,
			ExcludedIngredientIDs: append([]int(nil), item.ExcludedIngredientIDs...),
		})
	}
	return OrderResponse{
		ID: order.ID, ShopID: order.ShopID, Status: order.Status, Items: items,
		TotalCents: order.TotalCents, EstimatedTimeMin: order.EstimatedTimeMin,
		CreatedAt: order.CreatedAt, UpdatedAt: order.UpdatedAt,
	}
}
