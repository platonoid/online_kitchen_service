package services

import (
	"kitchen/internal/domain"
	"kitchen/internal/repository"
)

func validateExcluded(dish domain.Dish, excluded []int) error {
	allowed := make(map[int]bool, len(dish.Ingredients))
	for _, ingredient := range dish.Ingredients {
		allowed[ingredient.ID] = ingredient.Removable
	}
	for _, id := range excluded {
		if removable, ok := allowed[id]; !ok || !removable {
			return ErrIngredientNotRemovable
		}
	}
	return nil
}

func recalculateAll(order *domain.Order, repo repository.Repository) {
	order.TotalCents = 0
	maxPreparation := 0
	shop, _ := repo.GetShop(order.ShopID)
	for _, item := range order.Items {
		order.TotalCents += item.Quantity * item.UnitPriceCents
		if dish, err := repo.GetDish(item.DishID); err == nil && dish.PreparationTimeMin > maxPreparation {
			maxPreparation = dish.PreparationTimeMin
		}
	}
	order.EstimatedTimeMin = shop.DeliveryTimeMin + maxPreparation
}

func validStatus(status string) bool {
	switch status {
	case domain.OrderStatusPending, domain.OrderStatusAccepted, domain.OrderStatusCooking, domain.OrderStatusReady, domain.OrderStatusDelivered, domain.OrderStatusCancelled:
		return true
	default:
		return false
	}
}
