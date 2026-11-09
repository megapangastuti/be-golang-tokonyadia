package models

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID          uuid.UUID `json:"id"`
	Fullname    string    `json:"fullname" binding:"required"`
	Birthdate   time.Time `json:"birthdate" binding:"required"`
	Address     string    `json:"address" binding:"required"`
	PhoneNumber string    `json:"phoneNumber" binding:"required"`
	Email       string    `json:"email" binding:"required,email"`
	Password    string    `json:"password,omitempty" binding:"required,min=6"`
	Role        string    `json:"role"` // "customer" or "admin"
}

// RegisterRequest intentionally excludes ID and Role. Public registration must
// never allow a caller to choose their own authorization level.
type RegisterRequest struct {
	Fullname    string    `json:"fullname" binding:"required,max=100"`
	Birthdate   time.Time `json:"birthdate" binding:"required"`
	Address     string    `json:"address" binding:"required,max=500"`
	PhoneNumber string    `json:"phoneNumber" binding:"required,max=30"`
	Email       string    `json:"email" binding:"required,email,max=254"`
	Password    string    `json:"password" binding:"required,min=8,max=72"`
}

type CustomerResponse struct {
	ID          uuid.UUID `json:"id"`
	Fullname    string    `json:"fullname"`
	Birthdate   time.Time `json:"birthdate"`
	Address     string    `json:"address"`
	PhoneNumber string    `json:"phoneNumber"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
}

// UpdateCustomerRequest contains fields administrators may change. Role and
// ID are deliberately excluded to avoid privilege and identity changes via a
// general update endpoint.
type UpdateCustomerRequest struct {
	Fullname    string    `json:"fullname" binding:"required,max=100"`
	Birthdate   time.Time `json:"birthdate" binding:"required"`
	Address     string    `json:"address" binding:"required,max=500"`
	PhoneNumber string    `json:"phoneNumber" binding:"required,max=30"`
	Email       string    `json:"email" binding:"required,email,max=254"`
	Password    string    `json:"password" binding:"omitempty,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,max=72"`
}

type LoginResponse struct {
	Token     string           `json:"token"`
	TokenType string           `json:"tokenType"`
	ExpiresIn int64            `json:"expiresIn"`
	Customer  CustomerResponse `json:"customer"`
}
