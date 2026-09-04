package domain

type Ingredient struct {
	ID        int    `json:"id"`
	DishID    int    `json:"dish_id"`
	Name      string `json:"name"`
	Removable bool   `json:"removable"`
}
