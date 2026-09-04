package repository

import "kitchen/internal/domain"

type Repository interface {
	ListShops() []domain.Shop
	GetShop(id int) (domain.Shop, error)
	ListDishes(shopID int) []domain.Dish
	GetDish(id int) (domain.Dish, error)
	CreateOrder(order domain.Order) (domain.Order, error)
	GetOrder(id int) (domain.Order, error)
	SaveOrder(order domain.Order) error
	NextItemID() int
}
