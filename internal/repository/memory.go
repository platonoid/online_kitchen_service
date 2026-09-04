package repository

import (
	"sync"

	"kitchen/internal/domain"
)

type InMemoryRepository struct {
	mu          sync.RWMutex
	shops       map[int]domain.Shop
	dishes      map[int]domain.Dish
	orders      map[int]domain.Order
	nextOrderID int
	nextItemID  int
}

func NewInMemoryRepository() *InMemoryRepository {
	r := &InMemoryRepository{
		shops:       make(map[int]domain.Shop),
		dishes:      make(map[int]domain.Dish),
		orders:      make(map[int]domain.Order),
		nextOrderID: 1,
		nextItemID:  1,
	}
	r.seed()
	return r
}
