package handlers

import (
	"errors"
	"net/http"

	"tokonyadia/internal/models"
	"tokonyadia/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductHandler struct {
	service service.ProductService
}

func NewProductHandler(service service.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product (Admin only)
// @Tags products
// @Accept json
// @Produce json
// @Param product body models.ProductRequest true "Product data"
// @Param Authorization header string true "Bearer token"
// @Success 201 {object} models.Product
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /products [post]
// @Security BearerAuth
func (h *ProductHandler) Create(c *gin.Context) {
	var req models.ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	product := models.Product{
		ProductName:  req.ProductName,
		Stock:        req.Stock,
		ProductPrice: req.ProductPrice,
	}
	if err := h.service.Create(&product); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// GetProduct godoc
// @Summary Get a product by ID
// @Description Get product details by ID
// @Tags products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} models.Product
// @Failure 404 {object} models.ErrorResponse
// @Router /products/{id} [get]
func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid product ID format",
		})
		return
	}

	product, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, product)
}

// GetAllProducts godoc
// @Summary Get all products
// @Description Get all products with pagination, search, and sorting
// @Tags products
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(10)
// @Param search query string false "Search by product name"
// @Param sortBy query string false "Sort by field (productName, stock, productPrice)"
// @Param sortDir query string false "Sort direction (asc, desc)" default(asc)
// @Success 200 {object} models.PaginatedResponse
// @Router /products [get]
func (h *ProductHandler) GetAll(c *gin.Context) {
	var params models.PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	response, err := h.service.GetAll(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateProduct godoc
// @Summary Update a product
// @Description Update product details (Admin only)
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body models.ProductRequest true "Product data"
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} models.Product
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /products/{id} [put]
// @Security BearerAuth
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid product ID format",
		})
		return
	}

	var req models.ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	product := models.Product{
		ID:           id,
		ProductName:  req.ProductName,
		Stock:        req.Stock,
		ProductPrice: req.ProductPrice,
	}
	if err := h.service.Update(&product); err != nil {
		writeProductError(c, err)
		return
	}

	c.JSON(http.StatusOK, product)
}

// DeleteProduct godoc
// @Summary Delete a product
// @Description Delete a product by ID (Admin only)
// @Tags products
// @Param id path string true "Product ID"
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /products/{id} [delete]
// @Security BearerAuth
func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid product ID format",
		})
		return
	}

	if err := h.service.Delete(id); err != nil {
		writeProductError(c, err)
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Product deleted successfully",
	})
}

func writeProductError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrProductNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "Product not found"})
		return
	}
	c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Unable to process product"})
}
