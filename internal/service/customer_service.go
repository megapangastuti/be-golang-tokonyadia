package service

import (
	"errors"
	"fmt"
	"strings"

	"tokonyadia/internal/models"
	"tokonyadia/internal/repository"
	"tokonyadia/internal/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type CustomerService interface {
	Register(req models.RegisterRequest) (*models.CustomerResponse, error)
	Login(req models.LoginRequest) (*models.LoginResponse, error)
	GetByID(id uuid.UUID) (*models.CustomerResponse, error)
	GetAll(params models.PaginationParams) (*models.PaginatedResponse, error)
	Update(customer *models.Customer) error
	Delete(id uuid.UUID) error
}

const (
	defaultPageSize = 10
	maxPageSize     = 100
	defaultSortDir  = "asc"
	defaultSortBy   = "fullname"
	defaultRole     = "customer"
)

type customerService struct {
	repo repository.CustomerRepository
}

func NewCustomerService(repo repository.CustomerRepository) CustomerService {
	return &customerService{
		repo: repo,
	}
}

// Register creates a new customer account.
func (s *customerService) Register(
	req models.RegisterRequest,
) (*models.CustomerResponse, error) {

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash customer password: %w", err)
	}

	customer := &models.Customer{
		Fullname:    strings.TrimSpace(req.Fullname),
		Birthdate:   req.Birthdate,
		Address:     strings.TrimSpace(req.Address),
		PhoneNumber: strings.TrimSpace(req.PhoneNumber),
		Email:       normalizeEmail(req.Email),
		Password:    hashedPassword,
		Role:        defaultRole,
	}

	if err := s.repo.Create(customer); err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			return nil, ErrEmailAlreadyRegistered
		}

		return nil, fmt.Errorf("create customer: %w", err)
	}

	return toCustomerResponse(customer), nil
}

// Login authenticates a customer and generates an access token.
func (s *customerService) Login(
	req models.LoginRequest,
) (*models.LoginResponse, error) {

	customer, err := s.repo.GetByEmail(normalizeEmail(req.Email))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf("get customer by email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(customer.Password),
		[]byte(req.Password),
	); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := utils.GenerateToken(
		customer.ID,
		customer.Email,
		customer.Role,
	)
	if err != nil {
		return nil, fmt.Errorf("generate authentication token: %w", err)
	}

	return &models.LoginResponse{
		Token:     token,
		TokenType: "Bearer",
		ExpiresIn: int64(utils.TokenTTL().Seconds()),
		Customer:  *toCustomerResponse(customer),
	}, nil
}

// GetByID retrieves a customer by UUID.
func (s *customerService) GetByID(
	id uuid.UUID,
) (*models.CustomerResponse, error) {

	customer, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCustomerNotFound
		}

		return nil, fmt.Errorf("get customer by ID: %w", err)
	}

	return toCustomerResponse(customer), nil
}

// GetAll retrieves customers with pagination, searching, and sorting.
func (s *customerService) GetAll(
	params models.PaginationParams,
) (*models.PaginatedResponse, error) {

	params = normalizePaginationParams(params)

	customers, total, err := s.repo.GetAll(params)
	if err != nil {
		return nil, fmt.Errorf("get all customers: %w", err)
	}

	responses := make([]models.CustomerResponse, 0, len(customers))

	for i := range customers {
		responses = append(
			responses,
			*toCustomerResponse(&customers[i]),
		)
	}

	totalPages := int(total / params.PageSize)
	if total%params.PageSize > 0 {
		totalPages++
	}

	return &models.PaginatedResponse{
		Data:       responses,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}, nil
}

// Update updates customer information.
func (s *customerService) Update(customer *models.Customer) error {
	if customer == nil {
		return errors.New("customer cannot be nil")
	}

	existing, err := s.repo.GetByID(customer.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrCustomerNotFound
		}

		return fmt.Errorf("get customer before update: %w", err)
	}

	// Normalize email.
	customer.Email = normalizeEmail(customer.Email)

	// Prevent role modification through general customer update.
	customer.Role = existing.Role

	// Preserve existing password if no new password is provided.
	if customer.Password == "" {
		customer.Password = existing.Password
	} else {
		hashedPassword, err := hashPassword(customer.Password)
		if err != nil {
			return fmt.Errorf("hash updated password: %w", err)
		}

		customer.Password = hashedPassword
	}

	if err := s.repo.Update(customer); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrCustomerNotFound
		}

		return fmt.Errorf("update customer: %w", err)
	}

	return nil
}

// Delete deletes a customer by UUID.
func (s *customerService) Delete(id uuid.UUID) error {
	if err := s.repo.Delete(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrCustomerNotFound
		}

		return fmt.Errorf("delete customer: %w", err)
	}

	return nil
}

// normalizeEmail removes whitespace and converts email to lowercase.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// hashPassword generates a bcrypt password hash.
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

// normalizePaginationParams applies pagination and sorting defaults.
func normalizePaginationParams(
	params models.PaginationParams,
) models.PaginationParams {

	if params.Page < 1 {
		params.Page = 1
	}

	if params.PageSize < 1 {
		params.PageSize = defaultPageSize
	}

	if params.PageSize > maxPageSize {
		params.PageSize = maxPageSize
	}

	params.SortDir = strings.ToLower(strings.TrimSpace(params.SortDir))
	if params.SortDir != "asc" && params.SortDir != "desc" {
		params.SortDir = defaultSortDir
	}

	params.SortBy = strings.ToLower(strings.TrimSpace(params.SortBy))

	switch params.SortBy {
	case "fullname", "email", "birthdate":
		// Allowed sorting fields.
	default:
		params.SortBy = defaultSortBy
	}

	return params
}

// toCustomerResponse maps Customer entity to CustomerResponse.
func toCustomerResponse(
	customer *models.Customer,
) *models.CustomerResponse {

	return &models.CustomerResponse{
		ID:          customer.ID,
		Fullname:    customer.Fullname,
		Birthdate:   customer.Birthdate,
		Address:     customer.Address,
		PhoneNumber: customer.PhoneNumber,
		Email:       customer.Email,
		Role:        customer.Role,
	}
}
