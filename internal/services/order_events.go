package services

import (
	"context"
	"fmt"

	"kitchen/internal/domain"
)

func (s *OrderService) HandleRestaurantEvent(ctx context.Context, event domain.OrderEvent) error {
	if event.Type != "order.accepted" && event.Type != "order.rejected" {
		return nil
	}
	order, err := s.repo.GetOrder(event.OrderID)
	if err != nil {
		return fmt.Errorf("get order %d: %w", event.OrderID, err)
	}
	if order.ShopID != event.ShopID {
		return fmt.Errorf("restaurant event shop %d does not match order shop %d", event.ShopID, order.ShopID)
	}
	if event.Type == "order.accepted" {
		order.Status = domain.OrderStatusAccepted
	} else {
		order.Status = domain.OrderStatusRejected
	}
	return s.repo.SaveOrder(order)
}
