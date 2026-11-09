package service

import "errors"

var (
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrCustomerNotFound       = errors.New("customer not found")
	ErrProductNotFound        = errors.New("product not found")
	ErrTransactionNotFound    = errors.New("transaction not found")
	ErrInsufficientStock      = errors.New("insufficient stock")
)
