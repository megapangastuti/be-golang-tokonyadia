package models

import "github.com/google/uuid"

type Product struct {
	ID           uuid.UUID `json:"id"`
	ProductName  string    `json:"productName" binding:"required"`
	Stock        int       `json:"stock" binding:"required,min=0"`
	ProductPrice float64   `json:"productPrice" binding:"required,min=0"`
}

// ProductRequest is the writable representation of a product. IDs are always
// assigned by the server and therefore must not be accepted from clients.
type ProductRequest struct {
	ProductName  string  `json:"productName" binding:"required,max=200"`
	Stock        int     `json:"stock" binding:"min=0"`
	ProductPrice float64 `json:"productPrice" binding:"gt=0"`
}
