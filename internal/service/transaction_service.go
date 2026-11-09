package service

import (
	"errors"
	"fmt"
	"tokonyadia/internal/models"
	"tokonyadia/internal/repository"

	"github.com/google/uuid"
)

type TransactionService interface {
	Create(customerID uuid.UUID, req models.CreateTransactionRequest) (*models.Transaction, error)
	GetByID(id uuid.UUID) (*models.Transaction, error)
	GetAll(params models.PaginationParams, customerID *uuid.UUID) (*models.PaginatedResponse, error)
}

type transactionService struct {
	transactionRepo repository.TransactionRepository
	productRepo     repository.ProductRepository
	customerRepo    repository.CustomerRepository
}

func NewTransactionService(
	transactionRepo repository.TransactionRepository,
	productRepo repository.ProductRepository,
	customerRepo repository.CustomerRepository,
) TransactionService {
	return &transactionService{
		transactionRepo: transactionRepo,
		productRepo:     productRepo,
		customerRepo:    customerRepo,
	}
}

func (s *transactionService) Create(customerID uuid.UUID, req models.CreateTransactionRequest) (*models.Transaction, error) {
	// Get customer
	customer, err := s.customerRepo.GetByID(customerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, fmt.Errorf("get customer: %w", err)
	}

	// Validate products and check stock
	transactionDetails := make([]models.TransactionDetail, 0, len(req.Items))
	stockDeltas := make(map[uuid.UUID]int, len(req.Items))
	for _, item := range req.Items {
		product, err := s.productRepo.GetByID(item.ProductID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, fmt.Errorf("%w: %s", ErrProductNotFound, item.ProductID)
			}
			return nil, fmt.Errorf("get product: %w", err)
		}

		// Create transaction detail
		detail := models.TransactionDetail{
			Product:   *product,
			Quantity:  item.Quantity,
			PriceSell: product.ProductPrice,
		}
		transactionDetails = append(transactionDetails, detail)
		stockDeltas[item.ProductID] -= item.Quantity
	}

	if err := s.productRepo.UpdateStocks(stockDeltas); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProductNotFound
		}
		if errors.Is(err, repository.ErrInsufficientStock) {
			return nil, ErrInsufficientStock
		}
		return nil, fmt.Errorf("update stock: %w", err)
	}

	// Create transaction
	transaction := &models.Transaction{
		Customer:          *customer,
		TransactionDetail: transactionDetails,
	}

	err = s.transactionRepo.Create(transaction)
	if err != nil {
		// The in-memory repository currently cannot fail after stock validation,
		// but compensate defensively if another implementation does.
		for id, delta := range stockDeltas {
			stockDeltas[id] = -delta
		}
		_ = s.productRepo.UpdateStocks(stockDeltas)
		return nil, fmt.Errorf("create transaction: %w", err)
	}

	return transaction, nil
}

func (s *transactionService) GetByID(id uuid.UUID) (*models.Transaction, error) {
	transaction, err := s.transactionRepo.GetByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrTransactionNotFound
	}
	return transaction, err
}

func (s *transactionService) GetAll(params models.PaginationParams, customerID *uuid.UUID) (*models.PaginatedResponse, error) {
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

	transactions, total, err := s.transactionRepo.GetAll(params, customerID)
	if err != nil {
		return nil, err
	}

	totalPages := total / params.PageSize
	if total%params.PageSize > 0 {
		totalPages++
	}

	return &models.PaginatedResponse{
		Data:       transactions,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}, nil
}
