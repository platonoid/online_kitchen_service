package domain

type OrderItem struct {
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
