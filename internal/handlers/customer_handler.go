package handlers

import (
	"errors"
	"net/http"

	"tokonyadia/internal/models"
	"tokonyadia/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CustomerHandler struct {
	service service.CustomerService
}

func NewCustomerHandler(svc service.CustomerService) *CustomerHandler {
	return &CustomerHandler{
		service: svc,
	}
}

// Register godoc
// @Summary Register a customer
// @Tags auth
// @Accept json
// @Produce json
// @Param customer body models.RegisterRequest true "Registration data"
// @Success 201 {object} models.CustomerResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Router /auth/register [post]
func (h *CustomerHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest,
			"invalid_request", "Invalid registration data")
		return
	}

	response, err := h.service.Register(req)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyRegistered) {
			respondError(c, http.StatusConflict,
				"email_already_registered",
				"An account with that email already exists")
			return
		}

		respondError(c, http.StatusInternalServerError,
			"registration_failed", "Unable to register customer")
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Login godoc
// @Summary Authenticate a customer
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/login [post]
func (h *CustomerHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest,
			"invalid_request", "Invalid login data")
		return
	}

	response, err := h.service.Login(req)
	if err != nil {
		// Invalid credentials should return 401.
		// Other service errors should return 500.
		if errors.Is(err, service.ErrInvalidCredentials) {
			respondError(c, http.StatusUnauthorized,
				"invalid_credentials", "Invalid email or password")
			return
		}

		respondError(c, http.StatusInternalServerError,
			"internal_error", "Unable to process login")
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetByID godoc
// @Summary Get a customer by ID
// @Tags customers
// @Produce json
// @Param id path string true "Customer ID"
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} models.CustomerResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /customers/{id} [get]
// @Security BearerAuth
func (h *CustomerHandler) GetByID(c *gin.Context) {
	id, ok := parseCustomerID(c)
	if !ok {
		return
	}

	customer, err := h.service.GetByID(id)
	if err != nil {
		handleCustomerServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, customer)
}

// GetAll godoc
// @Summary List customers
// @Tags customers
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(10)
// @Param search query string false "Search by fullname or email"
// @Param sortBy query string false "Sort by field (fullname, email, birthdate)"
// @Param sortDir query string false "Sort direction (asc, desc)" default(asc)
// @Success 200 {object} models.PaginatedResponse
// @Router /customers [get]
// @Security BearerAuth
func (h *CustomerHandler) GetAll(c *gin.Context) {
	var params models.PaginationParams

	if err := c.ShouldBindQuery(&params); err != nil {
		respondError(c, http.StatusBadRequest,
			"invalid_request", "Invalid pagination parameters")
		return
	}

	response, err := h.service.GetAll(params)
	if err != nil {
		respondError(c, http.StatusInternalServerError,
			"internal_error", "Unable to retrieve customers")
		return
	}

	c.JSON(http.StatusOK, response)
}

// Update godoc
// @Summary Update a customer
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param customer body models.UpdateCustomerRequest true "Customer data"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /customers/{id} [put]
// @Security BearerAuth
func (h *CustomerHandler) Update(c *gin.Context) {
	id, ok := parseCustomerID(c)
	if !ok {
		return
	}

	var req models.UpdateCustomerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest,
			"invalid_request", "Invalid customer data")
		return
	}

	customer := models.Customer{
		ID:          id,
		Fullname:    req.Fullname,
		Birthdate:   req.Birthdate,
		Address:     req.Address,
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
		Password:    req.Password,
	}

	if err := h.service.Update(&customer); err != nil {
		handleCustomerServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Customer updated successfully",
	})
}

// Delete godoc
// @Summary Delete a customer
// @Tags customers
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /customers/{id} [delete]
// @Security BearerAuth
func (h *CustomerHandler) Delete(c *gin.Context) {
	id, ok := parseCustomerID(c)
	if !ok {
		return
	}

	if err := h.service.Delete(id); err != nil {
		handleCustomerServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Customer deleted successfully",
	})
}

// parseCustomerID parses the customer UUID from the URL parameter.
func parseCustomerID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest,
			"invalid_id", "Invalid customer ID format")
		return uuid.Nil, false
	}

	return id, true
}

// respondError sends a standardized error response.
func respondError(
	c *gin.Context,
	statusCode int,
	errorCode string,
	message string,
) {
	c.JSON(statusCode, models.ErrorResponse{
		Error:   errorCode,
		Message: message,
	})
}

// handleCustomerServiceError maps service errors to HTTP responses.
func handleCustomerServiceError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrCustomerNotFound) {
		respondError(c, http.StatusNotFound,
			"not_found", "Customer not found")
		return
	}

	respondError(c, http.StatusInternalServerError,
		"internal_error", "An unexpected error occurred")
}
