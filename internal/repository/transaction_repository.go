package repository

import (
	"sort"
	"strings"
	"sync"
	"time"

	"tokonyadia/internal/models"

	"github.com/google/uuid"
)

type TransactionRepository interface {
	Create(transaction *models.Transaction) error
	GetByID(id uuid.UUID) (*models.Transaction, error)
	GetAll(params models.PaginationParams, customerID *uuid.UUID) ([]models.Transaction, int, error)
	GetByCustomerID(customerID uuid.UUID) ([]models.Transaction, error)
}

type transactionRepository struct {
	transactions []models.Transaction
	mu           sync.RWMutex
}

func NewTransactionRepository() TransactionRepository {
	return &transactionRepository{
		transactions: []models.Transaction{},
	}
}

func (r *transactionRepository) Create(transaction *models.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	transaction.ID = uuid.New()
	transaction.TransactionDate = time.Now()

	// Generate IDs for transaction details
	for i := range transaction.TransactionDetail {
		transaction.TransactionDetail[i].ID = uuid.New()
	}

	r.transactions = append(r.transactions, *transaction)
	return nil
}

func (r *transactionRepository) GetByID(id uuid.UUID) (*models.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, t := range r.transactions {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, ErrNotFound
}

func (r *transactionRepository) GetAll(params models.PaginationParams, customerID *uuid.UUID) ([]models.Transaction, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Filter by customerID if provided (for non-admin users)
	filtered := make([]models.Transaction, 0)
	for _, t := range r.transactions {
		// Filter by customer if customerID is provided
		if customerID != nil && t.Customer.ID != *customerID {
			continue
		}

		// Filter by search (customer name or email)
		if params.Search == "" ||
			strings.Contains(strings.ToLower(t.Customer.Fullname), strings.ToLower(params.Search)) ||
			strings.Contains(strings.ToLower(t.Customer.Email), strings.ToLower(params.Search)) {
			filtered = append(filtered, t)
		}
	}

	// Sort
	if params.SortBy != "" {
		sort.Slice(filtered, func(i, j int) bool {
			switch params.SortBy {
			case "transactionDate":
				if params.SortDir == "desc" {
					return filtered[i].TransactionDate.After(filtered[j].TransactionDate)
				}
				return filtered[i].TransactionDate.Before(filtered[j].TransactionDate)
			case "customerName":
				if params.SortDir == "desc" {
					return filtered[i].Customer.Fullname > filtered[j].Customer.Fullname
				}
				return filtered[i].Customer.Fullname < filtered[j].Customer.Fullname
			default:
				// Default sort by date descending
				return filtered[i].TransactionDate.After(filtered[j].TransactionDate)
			}
		})
	} else {
		// Default sort by date descending
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].TransactionDate.After(filtered[j].TransactionDate)
		})
	}

	total := len(filtered)

	// Pagination
	start := (params.Page - 1) * params.PageSize
	end := start + params.PageSize

	if start > total {
		return []models.Transaction{}, total, nil
	}
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

func (r *transactionRepository) GetByCustomerID(customerID uuid.UUID) ([]models.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]models.Transaction, 0)
	for _, t := range r.transactions {
		if t.Customer.ID == customerID {
			result = append(result, t)
		}
	}
	return result, nil
}
