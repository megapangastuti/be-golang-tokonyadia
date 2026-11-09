package models

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID                uuid.UUID           `json:"id"`
	TransactionDate   time.Time           `json:"transactionDate"`
	Customer          Customer            `json:"customer"`
	TransactionDetail []TransactionDetail `json:"transactionDetail"`
}

type TransactionDetail struct {
	ID          uuid.UUID    `json:"id"`
	Product     Product      `json:"product"`
	Quantity    int          `json:"quantity" binding:"required,min=1"`
	PriceSell   float64      `json:"priceSell"`
	Transaction *Transaction `json:"-"` // Avoid circular reference in JSON
}

type CreateTransactionRequest struct {
	Items []TransactionItem `json:"items" binding:"required,min=1"`
}

type TransactionItem struct {
	ProductID uuid.UUID `json:"productId" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,min=1"`
}
