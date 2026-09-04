package services

import (
	"sort"
	"strings"

	"kitchen/internal/domain"
)

func (s *CatalogService) ListShops() []domain.Shop {
	return s.repo.ListShops()
}

func (s *CatalogService) GetShop(id int) (domain.Shop, error) {
	return s.repo.GetShop(id)
}

func (s *CatalogService) GetDish(id int) (domain.Dish, error) {
	return s.repo.GetDish(id)
}

func (s *CatalogService) ListDishes(shopID int, filter DishFilter) ([]domain.Dish, int, error) {
	if shopID != 0 {
		if _, err := s.repo.GetShop(shopID); err != nil {
			return nil, 0, err
		}
	}
	dishes := s.repo.ListDishes(shopID)
	query := strings.ToLower(strings.TrimSpace(filter.Query))
	category := strings.ToLower(strings.TrimSpace(filter.Category))
	filtered := make([]domain.Dish, 0, len(dishes))
	for _, dish := range dishes {
		if query != "" && !strings.Contains(strings.ToLower(dish.Name), query) && !strings.Contains(strings.ToLower(dish.Description), query) {
			continue
		}
		if category != "" && strings.ToLower(dish.Category) != category {
			continue
		}
		filtered = append(filtered, dish)
	}
	switch strings.ToLower(filter.Sort) {
	case "price", "price_asc":
		sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].PriceCents < filtered[j].PriceCents })
	case "price_desc":
		sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].PriceCents > filtered[j].PriceCents })
	case "name", "name_asc":
		sort.SliceStable(filtered, func(i, j int) bool { return strings.ToLower(filtered[i].Name) < strings.ToLower(filtered[j].Name) })
	case "name_desc":
		sort.SliceStable(filtered, func(i, j int) bool { return strings.ToLower(filtered[i].Name) > strings.ToLower(filtered[j].Name) })
	default:
		sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].ID < filtered[j].ID })
	}
	total := len(filtered)
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start >= len(filtered) {
		return []domain.Dish{}, total, nil
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], total, nil
}
