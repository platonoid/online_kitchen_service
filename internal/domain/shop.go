package domain

type Shop struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	City            string `json:"city"`
	Cuisine         string `json:"cuisine"`
	DeliveryTimeMin int    `json:"delivery_time_min"`
	IsActive        bool   `json:"is_active"`
}
