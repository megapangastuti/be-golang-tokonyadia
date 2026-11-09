package repository

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"tokonyadia/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type CustomerRepository interface {
	Create(customer *models.Customer) error
	GetByID(id uuid.UUID) (*models.Customer, error)
	GetByEmail(email string) (*models.Customer, error)
	GetAll(params models.PaginationParams) ([]models.Customer, int, error)
	Update(customer *models.Customer) error
	Delete(id uuid.UUID) error
}

type customerRepository struct {
	customers []models.Customer
	mu        sync.RWMutex
}

var (
	ErrNotFound           = errors.New("record not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInsufficientStock  = errors.New("insufficient stock")
)

// NewCustomerRepository initializes an in-memory customer repository.
func NewCustomerRepository() (CustomerRepository, error) {
	customers, err := getDefaultCustomers()
	if err != nil {
		return nil, fmt.Errorf("initialize default customers: %w", err)
	}

	return &customerRepository{
		customers: customers,
	}, nil
}

// getDefaultCustomers returns initial customer data.
func getDefaultCustomers() ([]models.Customer, error) {
	customerPassword, err := hashSeedPassword("password123")
	if err != nil {
		return nil, err
	}

	adminPassword, err := hashSeedPassword("admin123")
	if err != nil {
		return nil, err
	}

	return []models.Customer{
		{
			ID:          uuid.MustParse("650e8400-e29b-41d4-a716-446655440001"),
			Fullname:    "Admin User",
			Birthdate:   time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			Address:     "Jl. Admin No. 1, Jakarta",
			PhoneNumber: "081234567890",
			Email:       "admin@tokonyadia.com",
			Password:    adminPassword,
			Role:        "admin",
		},
		{
			ID:          uuid.MustParse("650e8400-e29b-41d4-a716-446655440002"),
			Fullname:    "Budi Santoso",
			Birthdate:   time.Date(1995, 5, 15, 0, 0, 0, 0, time.UTC),
			Address:     "Jl. Merdeka No. 123, Surabaya",
			PhoneNumber: "081234567891",
			Email:       "budi@email.com",
			Password:    customerPassword,
			Role:        "customer",
		},
		{
			ID:          uuid.MustParse("650e8400-e29b-41d4-a716-446655440003"),
			Fullname:    "Siti Nurhaliza",
			Birthdate:   time.Date(1998, 8, 20, 0, 0, 0, 0, time.UTC),
			Address:     "Jl. Sudirman No. 456, Bandung",
			PhoneNumber: "081234567892",
			Email:       "siti@email.com",
			Password:    customerPassword,
			Role:        "customer",
		},
	}, nil
}

// Create inserts a new customer.
func (r *customerRepository) Create(customer *models.Customer) error {
	if customer == nil {
		return errors.New("customer cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	email := normalizeEmail(customer.Email)

	if r.emailExists(email, uuid.Nil) {
		return ErrEmailAlreadyExists
	}

	customer.Email = email

	if customer.ID == uuid.Nil {
		customer.ID = uuid.New()
	}

	if customer.Role == "" {
		customer.Role = "customer"
	}

	r.customers = append(r.customers, *customer)

	return nil
}

// GetByID retrieves a customer by UUID.
func (r *customerRepository) GetByID(id uuid.UUID) (*models.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for i := range r.customers {
		if r.customers[i].ID == id {
			customer := r.customers[i]
			return &customer, nil
		}
	}

	return nil, ErrNotFound
}

// GetByEmail retrieves a customer by email.
func (r *customerRepository) GetByEmail(email string) (*models.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	email = normalizeEmail(email)

	for i := range r.customers {
		if r.customers[i].Email == email {
			customer := r.customers[i]
			return &customer, nil
		}
	}

	return nil, ErrNotFound
}

// GetAll retrieves customers with search, sorting, and pagination.
func (r *customerRepository) GetAll(
	params models.PaginationParams,
) ([]models.Customer, int, error) {

	// Take a snapshot under read lock.
	r.mu.RLock()
	customers := make([]models.Customer, len(r.customers))
	copy(customers, r.customers)
	r.mu.RUnlock()

	// Normalize parameters.
	params = normalizeRepositoryPagination(params)

	// Filter.
	filtered := filterCustomers(customers, params.Search)

	total := len(filtered)

	// Sort.
	sortCustomers(filtered, params.SortBy, params.SortDir)

	// Paginate.
	paginated := paginateCustomers(filtered, params.Page, params.PageSize)

	return paginated, total, nil
}

// Update replaces an existing customer.
func (r *customerRepository) Update(customer *models.Customer) error {
	if customer == nil {
		return errors.New("customer cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.customers {
		if r.customers[i].ID != customer.ID {
			continue
		}

		email := normalizeEmail(customer.Email)

		if r.emailExists(email, customer.ID) {
			return ErrEmailAlreadyExists
		}

		customer.Email = email
		r.customers[i] = *customer

		return nil
	}

	return ErrNotFound
}

// Delete removes a customer by UUID.
func (r *customerRepository) Delete(id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.customers {
		if r.customers[i].ID != id {
			continue
		}

		r.customers = append(
			r.customers[:i],
			r.customers[i+1:]...,
		)

		return nil
	}

	return ErrNotFound
}

// emailExists checks whether an email is already registered.
// excludeID allows the current customer to keep their own email.
func (r *customerRepository) emailExists(
	email string,
	excludeID uuid.UUID,
) bool {

	for i := range r.customers {
		if r.customers[i].ID == excludeID {
			continue
		}

		if normalizeEmail(r.customers[i].Email) == email {
			return true
		}
	}

	return false
}

// filterCustomers filters customers by fullname or email.
func filterCustomers(
	customers []models.Customer,
	search string,
) []models.Customer {

	search = strings.ToLower(strings.TrimSpace(search))

	if search == "" {
		return customers
	}

	filtered := make([]models.Customer, 0, len(customers))

	for _, customer := range customers {
		fullname := strings.ToLower(customer.Fullname)
		email := strings.ToLower(customer.Email)

		if strings.Contains(fullname, search) ||
			strings.Contains(email, search) {
			filtered = append(filtered, customer)
		}
	}

	return filtered
}

// sortCustomers sorts customers by an allowed field.
func sortCustomers(
	customers []models.Customer,
	sortBy string,
	sortDir string,
) {
	desc := strings.EqualFold(sortDir, "desc")

	sort.Slice(customers, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "email":
			less = strings.ToLower(customers[i].Email) <
				strings.ToLower(customers[j].Email)

		case "birthdate":
			less = customers[i].Birthdate.Before(customers[j].Birthdate)

		default:
			less = strings.ToLower(customers[i].Fullname) <
				strings.ToLower(customers[j].Fullname)
		}

		if desc {
			return !less && !customerSortEqual(customers[i], customers[j], sortBy)
		}

		return less
	})
}

// customerSortEqual checks equality of the selected sorting field.
func customerSortEqual(
	a models.Customer,
	b models.Customer,
	sortBy string,
) bool {
	switch sortBy {
	case "email":
		return strings.EqualFold(a.Email, b.Email)

	case "birthdate":
		return a.Birthdate.Equal(b.Birthdate)

	default:
		return strings.EqualFold(a.Fullname, b.Fullname)
	}
}

// paginateCustomers returns a page of customers safely.
func paginateCustomers(
	customers []models.Customer,
	page int,
	pageSize int,
) []models.Customer {

	total := len(customers)

	if total == 0 {
		return []models.Customer{}
	}

	if page > (total-1)/pageSize {
		return []models.Customer{}
	}

	start := (page - 1) * pageSize
	end := start + pageSize

	if end > total {
		end = total
	}

	return customers[start:end]
}

// normalizeRepositoryPagination applies safe defaults.
func normalizeRepositoryPagination(
	params models.PaginationParams,
) models.PaginationParams {

	if params.Page < 1 {
		params.Page = 1
	}

	if params.PageSize < 1 {
		params.PageSize = 10
	}

	if params.PageSize > 100 {
		params.PageSize = 100
	}

	params.SortBy = strings.ToLower(strings.TrimSpace(params.SortBy))

	switch params.SortBy {
	case "fullname", "email", "birthdate":
	default:
		params.SortBy = "fullname"
	}

	params.SortDir = strings.ToLower(strings.TrimSpace(params.SortDir))

	if params.SortDir != "asc" && params.SortDir != "desc" {
		params.SortDir = "asc"
	}

	return params
}

// normalizeEmail normalizes email values.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// hashSeedPassword hashes default seed passwords.
func hashSeedPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", fmt.Errorf("hash seed password: %w", err)
	}

	return string(hashed), nil
}
