package repository

import (
	"sort"

	"kitchen/internal/domain"
)

func (r *InMemoryRepository) ListShops() []domain.Shop {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]domain.Shop, 0, len(r.shops))
	for _, shop := range r.shops {
		if shop.IsActive {
			result = append(result, shop)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func (r *InMemoryRepository) GetShop(id int) (domain.Shop, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	shop, ok := r.shops[id]
	if !ok || !shop.IsActive {
		return domain.Shop{}, ErrNotFound
	}
	return shop, nil
}

func (r *InMemoryRepository) ListDishes(shopID int) []domain.Dish {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]domain.Dish, 0)
	for _, dish := range r.dishes {
		if (shopID == 0 || dish.ShopID == shopID) && dish.IsAvailable {
			result = append(result, cloneDish(dish))
		}
	}
	return result
}

func (r *InMemoryRepository) GetDish(id int) (domain.Dish, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	dish, ok := r.dishes[id]
	if !ok || !dish.IsAvailable {
		return domain.Dish{}, ErrNotFound
	}
	return cloneDish(dish), nil
}

func cloneDish(dish domain.Dish) domain.Dish {
	dish.Ingredients = append([]domain.Ingredient(nil), dish.Ingredients...)
	return dish
}
