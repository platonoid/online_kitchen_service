package db

type Shop struct {
	ID              int
	Name            string
	City            string
	Cuisine         string
	DeliveryTimeMin int
	IsActive        bool
}

type Dish struct {
	ID                 int
	ShopID             int
	Name               string
	Description        string
	Category           string
	PriceCents         int
	PreparationTimeMin int
	IsAvailable        bool
}

type Order struct {
	ID               int
	ShopID           int
	Status           string
	TotalCents       int
	EstimatedTimeMin int
}
