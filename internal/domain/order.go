package domain

import "time"

type Order struct {
	ID               int         `json:"id"`
	ShopID           int         `json:"shop_id"`
	Status           string      `json:"status"`
	Items            []OrderItem `json:"items"`
	TotalCents       int         `json:"total_cents"`
	EstimatedTimeMin int         `json:"estimated_time_min"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}
