package handlers

import (
	"errors"
	"net/http"

	"tokonyadia/internal/models"
	"tokonyadia/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TransactionHandler struct {
	service service.TransactionService
}

func NewTransactionHandler(service service.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

// CreateTransaction godoc
// @Summary Create a new transaction
// @Description Create a new transaction (purchase products)
// @Tags transactions
// @Accept json
// @Produce json
// @Param transaction body models.CreateTransactionRequest true "Transaction data"
// @Param Authorization header string true "Bearer token"
// @Success 201 {object} models.Transaction
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /transactions [post]
// @Security BearerAuth
func (h *TransactionHandler) Create(c *gin.Context) {
	customerID, ok := authenticatedCustomerID(c)
	if !ok {
		return
	}

	var req models.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	transaction, err := h.service.Create(customerID, req)
	if err != nil {
		writeTransactionError(c, err)
		return
	}

	c.JSON(http.StatusCreated, transaction)
}

// GetTransaction godoc
// @Summary Get a transaction by ID
// @Description Get transaction details by ID
// @Tags transactions
// @Produce json
// @Param id path string true "Transaction ID"
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} models.Transaction
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /transactions/{id} [get]
// @Security BearerAuth
func (h *TransactionHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid transaction ID format",
		})
		return
	}

	transaction, err := h.service.GetByID(id)
	if err != nil {
		writeTransactionError(c, err)
		return
	}

	// Check if user is admin or the owner of the transaction
	if c.GetString("role") != "admin" {
		customerID, ok := authenticatedCustomerID(c)
		if !ok {
			return
		}
		if transaction.Customer.ID != customerID {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have permission to view this transaction",
			})
			return
		}
	}

	c.JSON(http.StatusOK, transaction)
}

// GetAllTransactions godoc
// @Summary Get all transactions
// @Description Get all transactions with pagination, search, and sorting. Admin can see all transactions, customers can only see their own
// @Tags transactions
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(10)
// @Param search query string false "Search by customer name or email"
// @Param sortBy query string false "Sort by field (transactionDate, customerName)"
// @Param sortDir query string false "Sort direction (asc, desc)" default(asc)
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} models.PaginatedResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /transactions [get]
// @Security BearerAuth
func (h *TransactionHandler) GetAll(c *gin.Context) {
	var params models.PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	role := c.GetString("role")
	var customerID *uuid.UUID

	// If not admin, filter by customer ID
	if role != "admin" {
		id, ok := authenticatedCustomerID(c)
		if !ok {
			return
		}
		customerID = &id
	}

	response, err := h.service.GetAll(params, customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetMyTransactions godoc
// @Summary Get my transaction history
// @Description Get transaction history for the logged-in customer
// @Tags transactions
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(10)
// @Param sortBy query string false "Sort by field (transactionDate)"
// @Param sortDir query string false "Sort direction (asc, desc)" default(desc)
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} models.PaginatedResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /transactions/me [get]
// @Security BearerAuth
func (h *TransactionHandler) GetMyTransactions(c *gin.Context) {
	var params models.PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Get customer ID from token
	custID, ok := authenticatedCustomerID(c)
	if !ok {
		return
	}

	response, err := h.service.GetAll(params, &custID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func authenticatedCustomerID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get("customerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized", Message: "Customer ID not found in token"})
		return uuid.Nil, false
	}
	id, ok := value.(uuid.UUID)
	if !ok || id == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized", Message: "Invalid customer ID in token"})
		return uuid.Nil, false
	}
	return id, true
}

func writeTransactionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrTransactionNotFound):
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "Transaction not found"})
	case errors.Is(err, service.ErrCustomerNotFound), errors.Is(err, service.ErrProductNotFound), errors.Is(err, service.ErrInsufficientStock):
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "transaction_failed", Message: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Unable to process transaction"})
	}
}

// GetCustomerTransactions godoc
// @Summary Get transactions by customer ID
// @Description Get transaction history for a specific customer (Admin only)
// @Tags transactions
// @Produce json
// @Param customerId path string true "Customer ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(10)
// @Param sortBy query string false "Sort by field (transactionDate)"
// @Param sortDir query string false "Sort direction (asc, desc)" default(desc)
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} models.PaginatedResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /transactions/customer/{customerId} [get]
// @Security BearerAuth
func (h *TransactionHandler) GetCustomerTransactions(c *gin.Context) {
	var params models.PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Parse customer ID from path parameter
	customerID, err := uuid.Parse(c.Param("customerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid customer ID format",
		})
		return
	}

	response, err := h.service.GetAll(params, &customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
