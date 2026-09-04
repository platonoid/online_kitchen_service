package domain

import "time"

type OrderEvent struct {
	Type      string    `json:"type"`
	OrderID   int       `json:"order_id"`
	ItemID    int       `json:"item_id,omitempty"`
	DishID    int       `json:"dish_id,omitempty"`
	ShopID    int       `json:"shop_id"`
	Order     Order     `json:"order"`
	Reason    string    `json:"reason,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
