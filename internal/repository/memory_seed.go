package repository

import "kitchen/internal/domain"

func (r *InMemoryRepository) seed() {
	r.shops[1] = domain.Shop{ID: 1, Name: "Chai Corner", City: "Moscow", Cuisine: "Indian", DeliveryTimeMin: 25, IsActive: true}
	r.shops[2] = domain.Shop{ID: 2, Name: "PastaLab", City: "Moscow", Cuisine: "Italian", DeliveryTimeMin: 30, IsActive: true}
	r.shops[3] = domain.Shop{ID: 3, Name: "Sushi Metro", City: "Moscow", Cuisine: "Japanese", DeliveryTimeMin: 22, IsActive: true}
	r.addDish(domain.Dish{ID: 1, ShopID: 1, Name: "Masala Chai", Category: "drinks", Description: "Black tea with cardamom and milk.", PriceCents: 250, PreparationTimeMin: 5, IsAvailable: true, Ingredients: []domain.Ingredient{{ID: 1, DishID: 1, Name: "Sugar", Removable: true}, {ID: 2, DishID: 1, Name: "Milk", Removable: true}}})
	r.addDish(domain.Dish{ID: 2, ShopID: 1, Name: "Chicken Biryani", Category: "main", Description: "Fragrant rice with chicken and herbs.", PriceCents: 420, PreparationTimeMin: 20, IsAvailable: true, Ingredients: []domain.Ingredient{{ID: 3, DishID: 2, Name: "Cilantro", Removable: true}, {ID: 4, DishID: 2, Name: "Chicken", Removable: false}}})
	r.addDish(domain.Dish{ID: 3, ShopID: 1, Name: "Paneer Wrap", Category: "main", Description: "Grilled paneer with salad and house sauce.", PriceCents: 390, PreparationTimeMin: 15, IsAvailable: true, Ingredients: []domain.Ingredient{{ID: 5, DishID: 3, Name: "Onion", Removable: true}}})
	r.addDish(domain.Dish{ID: 4, ShopID: 2, Name: "Margherita Pizza", Category: "pizza", Description: "Tomato sauce, mozzarella and basil.", PriceCents: 540, PreparationTimeMin: 20, IsAvailable: true, Ingredients: []domain.Ingredient{{ID: 6, DishID: 4, Name: "Basil", Removable: true}}})
	r.addDish(domain.Dish{ID: 5, ShopID: 2, Name: "Spaghetti Carbonara", Category: "pasta", Description: "Creamy pasta with bacon and parmesan.", PriceCents: 620, PreparationTimeMin: 20, IsAvailable: true})
	r.addDish(domain.Dish{ID: 6, ShopID: 2, Name: "Caesar Salad", Category: "salad", Description: "Romaine, chicken, parmesan and dressing.", PriceCents: 430, PreparationTimeMin: 10, IsAvailable: true})
	r.addDish(domain.Dish{ID: 7, ShopID: 3, Name: "Salmon Roll Set", Category: "rolls", Description: "Classic salmon sushi set with miso soup.", PriceCents: 760, PreparationTimeMin: 15, IsAvailable: true})
	r.addDish(domain.Dish{ID: 8, ShopID: 3, Name: "Vegan Maki Box", Category: "maki", Description: "Fresh vegetables and avocado in rice rolls.", PriceCents: 490, PreparationTimeMin: 15, IsAvailable: true})
	r.addDish(domain.Dish{ID: 9, ShopID: 3, Name: "Edamame Bowl", Category: "snacks", Description: "Warm edamame with sesame and chili.", PriceCents: 310, PreparationTimeMin: 8, IsAvailable: true})
}

func (r *InMemoryRepository) addDish(dish domain.Dish) {
	r.dishes[dish.ID] = dish
}
