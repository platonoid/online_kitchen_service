package services

import (
	"context"
	"testing"

	"kitchen/internal/domain"
	"kitchen/internal/repository"
)

type testPublisher struct {
	events []domain.OrderEvent
}

func (p *testPublisher) Publish(_ context.Context, event domain.OrderEvent) error {
	p.events = append(p.events, event)
	return nil
}

func TestOrderServiceItemsAndExclusions(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	publisher := &testPublisher{}
	service := NewOrderService(repo, publisher)
	order, err := service.CreateOrder(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	order, err = service.AddItem(context.Background(), order.ID, AddItemInput{DishID: 1, Quantity: 2, ExcludedIngredientIDs: []int{1}})
	if err != nil {
		t.Fatal(err)
	}
	if order.TotalCents != 500 || len(order.Items) != 1 {
		t.Fatalf("unexpected order: %+v", order)
	}
	if _, err = service.UpdateItem(context.Background(), order.ID, order.Items[0].ID, UpdateItemInput{Quantity: 3, ExcludedIngredientIDs: []int{4}}); err != ErrIngredientNotRemovable {
		t.Fatalf("expected non-removable ingredient error, got %v", err)
	}
	order, err = service.DeleteItem(context.Background(), order.ID, order.Items[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if order.TotalCents != 0 || len(order.Items) != 0 {
		t.Fatalf("expected empty order: %+v", order)
	}
	if len(publisher.events) != 2 || publisher.events[1].Type != "order.item.added" {
		t.Fatalf("unexpected events: %+v", publisher.events)
	}
}

func TestCatalogFilteringSortingAndPagination(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	service := NewCatalogService(repo)
	dishes, total, err := service.ListDishes(1, DishFilter{Category: "main", Sort: "price_desc", Page: 1, PageSize: 1})
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(dishes) != 1 || dishes[0].Name != "Chicken Biryani" {
		t.Fatalf("unexpected dishes: total=%d dishes=%+v", total, dishes)
	}
}
