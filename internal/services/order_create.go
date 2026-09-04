package services

import (
	"context"
	"fmt"
	"time"

	"kitchen/internal/domain"
)

func (s *OrderService) CreateOrder(ctx context.Context, shopID int) (domain.Order, error) {
	if _, err := s.repo.GetShop(shopID); err != nil {
		return domain.Order{}, err
	}
	order, err := s.repo.CreateOrder(domain.Order{ShopID: shopID, Status: domain.OrderStatusPending, Items: []domain.OrderItem{}})
	if err != nil {
		return domain.Order{}, err
	}
	if s.publisher != nil {
		if publishErr := s.publisher.Publish(ctx, domain.OrderEvent{Type: "order.created", OrderID: order.ID, ShopID: order.ShopID, Order: order, CreatedAt: time.Now().UTC()}); publishErr != nil {
			return domain.Order{}, fmt.Errorf("publish order event: %w", publishErr)
		}
	}
	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderID int) (domain.Order, error) {
	return s.repo.GetOrder(orderID)
}

func (s *OrderService) EstimateTime(ctx context.Context, orderID int) (int, error) {
	order, err := s.repo.GetOrder(orderID)
	if err != nil {
		return 0, err
	}
	shop, err := s.repo.GetShop(order.ShopID)
	if err != nil {
		return 0, err
	}
	maxPreparation := 0
	for _, item := range order.Items {
		dish, dishErr := s.repo.GetDish(item.DishID)
		if dishErr == nil && dish.PreparationTimeMin > maxPreparation {
			maxPreparation = dish.PreparationTimeMin
		}
	}
	return shop.DeliveryTimeMin + maxPreparation, nil
}
