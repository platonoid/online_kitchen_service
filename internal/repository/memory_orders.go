package repository

import (
	"time"

	"kitchen/internal/domain"
)

func (r *InMemoryRepository) CreateOrder(order domain.Order) (domain.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	shop, ok := r.shops[order.ShopID]
	if !ok || !shop.IsActive {
		return domain.Order{}, ErrNotFound
	}
	now := time.Now().UTC()
	order.ID = r.nextOrderID
	r.nextOrderID++
	order.CreatedAt = now
	order.UpdatedAt = now
	order.Items = []domain.OrderItem{}
	r.orders[order.ID] = cloneOrder(order)
	return cloneOrder(order), nil
}

func (r *InMemoryRepository) GetOrder(id int) (domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	order, ok := r.orders[id]
	if !ok {
		return domain.Order{}, ErrNotFound
	}
	return cloneOrder(order), nil
}

func (r *InMemoryRepository) SaveOrder(order domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.orders[order.ID]; !ok {
		return ErrNotFound
	}
	order.UpdatedAt = time.Now().UTC()
	r.orders[order.ID] = cloneOrder(order)
	return nil
}

func (r *InMemoryRepository) NextItemID() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := r.nextItemID
	r.nextItemID++
	return id
}

func cloneOrder(order domain.Order) domain.Order {
	order.Items = append([]domain.OrderItem(nil), order.Items...)
	for i := range order.Items {
		order.Items[i].ExcludedIngredientIDs = append([]int(nil), order.Items[i].ExcludedIngredientIDs...)
	}
	return order
}

type MemoryRepository = InMemoryRepository

func NewMemoryRepository() *InMemoryRepository {
	return NewInMemoryRepository()
}
