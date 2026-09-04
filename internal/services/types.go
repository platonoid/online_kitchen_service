package services

import (
	"context"

	"kitchen/internal/domain"
	"kitchen/internal/repository"
)

type DishFilter struct {
	Query    string
	Category string
	Sort     string
	Page     int
	PageSize int
}

type EventPublisher interface {
	Publish(context.Context, domain.OrderEvent) error
}

type EventHandler interface {
	Handle(context.Context, domain.OrderEvent) error
}

type AddItemInput struct {
	DishID                int
	Quantity              int
	ExcludedIngredientIDs []int
}

type UpdateItemInput struct {
	Quantity              int
	ExcludedIngredientIDs []int
}

type CatalogService struct {
	repo repository.Repository
}

type OrderService struct {
	repo      repository.Repository
	publisher EventPublisher
}

func NewCatalogService(repo repository.Repository) *CatalogService {
	return &CatalogService{repo: repo}
}

func NewOrderService(repo repository.Repository, publishers ...EventPublisher) *OrderService {
	var publisher EventPublisher
	if len(publishers) > 0 {
		publisher = publishers[0]
	}
	return &OrderService{repo: repo, publisher: publisher}
}
