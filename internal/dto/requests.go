package dto

type CreateOrderRequest struct {
	ShopID int `json:"shop_id"`
}

type AddOrderItemRequest struct {
	DishID                int   `json:"dish_id"`
	ItemID                int   `json:"item_id"`
	Quantity              int   `json:"quantity"`
	ExcludedIngredientIDs []int `json:"excluded_ingredient_ids"`
}

type UpdateOrderItemRequest struct {
	Quantity              int   `json:"quantity"`
	ExcludedIngredientIDs []int `json:"excluded_ingredient_ids"`
}

type DishListRequest struct {
	ShopID   int
	Type     string
	Query    string
	Sort     string
	Page     int
	PageSize int
}
