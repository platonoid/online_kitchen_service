package domain

type Dish struct {
	ID                 int          `json:"id"`
	ShopID             int          `json:"shop_id"`
	Name               string       `json:"name"`
	Description        string       `json:"description,omitempty"`
	Category           string       `json:"category"`
	PriceCents         int          `json:"price_cents"`
	PreparationTimeMin int          `json:"preparation_time_min"`
	IsAvailable        bool         `json:"is_available"`
	Ingredients        []Ingredient `json:"ingredients,omitempty"`
}
