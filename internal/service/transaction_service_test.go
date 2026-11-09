package service

import (
	"errors"
	"testing"

	"tokonyadia/internal/models"
	"tokonyadia/internal/repository"

	"github.com/google/uuid"
)

func TestCreateDoesNotPartiallyDeductStockForDuplicateProduct(t *testing.T) {
	productRepo := repository.NewProductRepository()
	transactionRepo := repository.NewTransactionRepository()
	customerRepo, err := repository.NewCustomerRepository()
	if err != nil {
		t.Fatalf("create customer repository: %v", err)
	}

	service := NewTransactionService(transactionRepo, productRepo, customerRepo)
	productID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	customerID := uuid.MustParse("650e8400-e29b-41d4-a716-446655440002")

	_, err = service.Create(customerID, models.CreateTransactionRequest{
		Items: []models.TransactionItem{
			{ProductID: productID, Quantity: 6},
			{ProductID: productID, Quantity: 6},
		},
	})
	if !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("Create() error = %v, want ErrInsufficientStock", err)
	}

	product, err := productRepo.GetByID(productID)
	if err != nil {
		t.Fatalf("get product: %v", err)
	}
	if product.Stock != 10 {
		t.Fatalf("stock = %d, want 10 after rejected transaction", product.Stock)
	}
}

func TestCreateDeductsAggregatedStockOnce(t *testing.T) {
	productRepo := repository.NewProductRepository()
	transactionRepo := repository.NewTransactionRepository()
	customerRepo, err := repository.NewCustomerRepository()
	if err != nil {
		t.Fatalf("create customer repository: %v", err)
	}

	service := NewTransactionService(transactionRepo, productRepo, customerRepo)
	productID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	customerID := uuid.MustParse("650e8400-e29b-41d4-a716-446655440002")

	transaction, err := service.Create(customerID, models.CreateTransactionRequest{
		Items: []models.TransactionItem{
			{ProductID: productID, Quantity: 3},
			{ProductID: productID, Quantity: 2},
		},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if transaction.ID == uuid.Nil {
		t.Fatal("created transaction has no ID")
	}

	product, err := productRepo.GetByID(productID)
	if err != nil {
		t.Fatalf("get product: %v", err)
	}
	if product.Stock != 5 {
		t.Fatalf("stock = %d, want 5", product.Stock)
	}
}
