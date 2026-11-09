package repository

import (
	"errors"
	"sort"
	"strings"
	"sync"

	"tokonyadia/internal/models"

	"github.com/google/uuid"
)

type ProductRepository interface {
	Create(product *models.Product) error
	GetByID(id uuid.UUID) (*models.Product, error)
	GetAll(params models.PaginationParams) ([]models.Product, int, error)
	Update(product *models.Product) error
	Delete(id uuid.UUID) error
	UpdateStock(id uuid.UUID, quantity int) error
	UpdateStocks(deltas map[uuid.UUID]int) error
}

type productRepository struct {
	products []models.Product
	mu       sync.RWMutex
}

func NewProductRepository() ProductRepository {
	repo := &productRepository{
		products: getDefaultProducts(),
	}
	return repo
}

func getDefaultProducts() []models.Product {
	return []models.Product{
		{
			ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			ProductName:  "Laptop ASUS ROG",
			Stock:        10,
			ProductPrice: 15000000,
		},
		{
			ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			ProductName:  "Mouse Logitech MX Master",
			Stock:        25,
			ProductPrice: 1200000,
		},
		{
			ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
			ProductName:  "Keyboard Mechanical",
			Stock:        15,
			ProductPrice: 850000,
		},
		{
			ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440004"),
			ProductName:  "Monitor LG 27 inch",
			Stock:        8,
			ProductPrice: 3500000,
		},
		{
			ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440005"),
			ProductName:  "Headset Gaming",
			Stock:        20,
			ProductPrice: 750000,
		},
	}
}

func (r *productRepository) Create(product *models.Product) error {
	if product == nil {
		return errors.New("product cannot be nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	product.ID = uuid.New()
	r.products = append(r.products, *product)
	return nil
}

func (r *productRepository) GetByID(id uuid.UUID) (*models.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.products {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, ErrNotFound
}

func (r *productRepository) GetAll(params models.PaginationParams) ([]models.Product, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Filter by search
	filtered := make([]models.Product, 0)
	for _, p := range r.products {
		if params.Search == "" || strings.Contains(strings.ToLower(p.ProductName), strings.ToLower(params.Search)) {
			filtered = append(filtered, p)
		}
	}

	// Sort
	if params.SortBy != "" {
		sort.Slice(filtered, func(i, j int) bool {
			switch params.SortBy {
			case "productName":
				if params.SortDir == "desc" {
					return filtered[i].ProductName > filtered[j].ProductName
				}
				return filtered[i].ProductName < filtered[j].ProductName
			case "stock":
				if params.SortDir == "desc" {
					return filtered[i].Stock > filtered[j].Stock
				}
				return filtered[i].Stock < filtered[j].Stock
			case "productPrice":
				if params.SortDir == "desc" {
					return filtered[i].ProductPrice > filtered[j].ProductPrice
				}
				return filtered[i].ProductPrice < filtered[j].ProductPrice
			default:
				return true
			}
		})
	}

	total := len(filtered)

	// Pagination
	start := (params.Page - 1) * params.PageSize
	end := start + params.PageSize

	if start > total {
		return []models.Product{}, total, nil
	}
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

func (r *productRepository) Update(product *models.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, p := range r.products {
		if p.ID == product.ID {
			r.products[i] = *product
			return nil
		}
	}
	return ErrNotFound
}

func (r *productRepository) Delete(id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, p := range r.products {
		if p.ID == id {
			r.products = append(r.products[:i], r.products[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (r *productRepository) UpdateStock(id uuid.UUID, quantity int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, p := range r.products {
		if p.ID == id {
			newStock := p.Stock + quantity
			if newStock < 0 {
				return ErrInsufficientStock
			}
			r.products[i].Stock = newStock
			return nil
		}
	}
	return ErrNotFound
}

// UpdateStocks validates and applies all stock changes while holding one lock.
// This prevents partial deductions when an order contains repeated products or
// when one of several products has insufficient stock.
func (r *productRepository) UpdateStocks(deltas map[uuid.UUID]int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	indexes := make(map[uuid.UUID]int, len(deltas))
	for id := range deltas {
		for i := range r.products {
			if r.products[i].ID == id {
				indexes[id] = i
				break
			}
		}
		if _, ok := indexes[id]; !ok {
			return ErrNotFound
		}
	}

	for id, delta := range deltas {
		if r.products[indexes[id]].Stock+delta < 0 {
			return ErrInsufficientStock
		}
	}

	for id, delta := range deltas {
		r.products[indexes[id]].Stock += delta
	}

	return nil
}
