package service

import (
	"errors"
	"fmt"
	"tokonyadia/internal/models"
	"tokonyadia/internal/repository"

	"github.com/google/uuid"
)

type ProductService interface {
	Create(product *models.Product) error
	GetByID(id uuid.UUID) (*models.Product, error)
	GetAll(params models.PaginationParams) (*models.PaginatedResponse, error)
	Update(product *models.Product) error
	Delete(id uuid.UUID) error
}

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) Create(product *models.Product) error {
	if err := s.repo.Create(product); err != nil {
		return fmt.Errorf("create product: %w", err)
	}
	return nil
}

func (s *productService) GetByID(id uuid.UUID) (*models.Product, error) {
	product, err := s.repo.GetByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrProductNotFound
	}
	return product, err
}

func (s *productService) GetAll(params models.PaginationParams) (*models.PaginatedResponse, error) {
	// Set default values
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 10
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}
	if params.SortDir != "asc" && params.SortDir != "desc" {
		params.SortDir = "asc"
	}

	products, total, err := s.repo.GetAll(params)
	if err != nil {
		return nil, fmt.Errorf("get products: %w", err)
	}

	totalPages := total / params.PageSize
	if total%params.PageSize > 0 {
		totalPages++
	}

	return &models.PaginatedResponse{
		Data:       products,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}, nil
}

func (s *productService) Update(product *models.Product) error {
	// Check if product exists
	if product == nil {
		return errors.New("product cannot be nil")
	}
	if _, err := s.GetByID(product.ID); err != nil {
		return err
	}
	if err := s.repo.Update(product); err != nil {
		return fmt.Errorf("update product: %w", err)
	}
	return nil
}

func (s *productService) Delete(id uuid.UUID) error {
	// Check if product exists
	if _, err := s.GetByID(id); err != nil {
		return err
	}
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	return nil
}
