package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"kitchen/internal/domain"
	"kitchen/internal/repository"
)

var (
	ErrInvalidQuantity        = errors.New("quantity must be greater than zero")
	ErrIngredientNotRemovable = errors.New("ingredient cannot be excluded")
	ErrDishUnavailable        = errors.New("dish is unavailable")
	ErrOrderShopMismatch      = errors.New("dish belongs to another shop")
)

func (s *OrderService) AddItem(ctx context.Context, orderID int, input AddItemInput) (domain.Order, error) {
	if input.Quantity < 1 {
		return domain.Order{}, ErrInvalidQuantity
	}
	order, err := s.repo.GetOrder(orderID)
	if err != nil {
		return domain.Order{}, err
	}
	dish, err := s.repo.GetDish(input.DishID)
	if err != nil {
		return domain.Order{}, err
	}
	if dish.ShopID != order.ShopID {
		return domain.Order{}, ErrOrderShopMismatch
	}
	if err := validateExcluded(dish, input.ExcludedIngredientIDs); err != nil {
		return domain.Order{}, err
	}
	itemIndex := -1
	for i := range order.Items {
		if order.Items[i].DishID == input.DishID {
			itemIndex = i
			break
		}
	}
	if itemIndex >= 0 {
		order.Items[itemIndex].Quantity += input.Quantity
		order.Items[itemIndex].ExcludedIngredientIDs = append([]int(nil), input.ExcludedIngredientIDs...)
	} else {
		order.Items = append(order.Items, domain.OrderItem{ID: s.repo.NextItemID(), OrderID: order.ID, DishID: dish.ID, DishName: dish.Name, Quantity: input.Quantity, UnitPriceCents: dish.PriceCents, Status: "PENDING", ExcludedIngredientIDs: append([]int(nil), input.ExcludedIngredientIDs...)})
	}
	recalculateAll(&order, s.repo)
	if err := s.repo.SaveOrder(order); err != nil {
		return domain.Order{}, err
	}
	if s.publisher != nil {
		item := order.Items[len(order.Items)-1]
		if itemIndex >= 0 {
			item = order.Items[itemIndex]
		}
		if publishErr := s.publisher.Publish(ctx, domain.OrderEvent{Type: "order.item.added", OrderID: order.ID, ItemID: item.ID, DishID: item.DishID, ShopID: order.ShopID, Order: order, CreatedAt: time.Now().UTC()}); publishErr != nil {
			return domain.Order{}, fmt.Errorf("publish item event: %w", publishErr)
		}
	}
	return s.repo.GetOrder(order.ID)
}

func (s *OrderService) UpdateItem(ctx context.Context, orderID, itemID int, input UpdateItemInput) (domain.Order, error) {
	if input.Quantity < 1 {
		return domain.Order{}, ErrInvalidQuantity
	}
	order, err := s.repo.GetOrder(orderID)
	if err != nil {
		return domain.Order{}, err
	}
	for i := range order.Items {
		if order.Items[i].ID != itemID {
			continue
		}
		dish, dishErr := s.repo.GetDish(order.Items[i].DishID)
		if dishErr != nil {
			return domain.Order{}, dishErr
		}
		if err := validateExcluded(dish, input.ExcludedIngredientIDs); err != nil {
			return domain.Order{}, err
		}
		order.Items[i].Quantity = input.Quantity
		order.Items[i].ExcludedIngredientIDs = append([]int(nil), input.ExcludedIngredientIDs...)
		recalculateAll(&order, s.repo)
		if err := s.repo.SaveOrder(order); err != nil {
			return domain.Order{}, err
		}
		return s.repo.GetOrder(order.ID)
	}
	return domain.Order{}, repository.ErrNotFound
}

func (s *OrderService) DeleteItem(ctx context.Context, orderID, itemID int) (domain.Order, error) {
	order, err := s.repo.GetOrder(orderID)
	if err != nil {
		return domain.Order{}, err
	}
	for i := range order.Items {
		if order.Items[i].ID == itemID {
			order.Items = append(order.Items[:i], order.Items[i+1:]...)
			recalculateAll(&order, s.repo)
			if err := s.repo.SaveOrder(order); err != nil {
				return domain.Order{}, err
			}
			return s.repo.GetOrder(order.ID)
		}
	}
	return domain.Order{}, repository.ErrNotFound
}
