# Tokonyadia - E-commerce Backend API

Backend REST API untuk aplikasi e-commerce Tokonyadia yang dibangun dengan Golang, Gin Framework, dan JWT authentication.

## Fitur

- **Management Produk**: CRUD produk dengan pagination, searching, dan sorting
- **Management Customer**: CRUD customer dengan pagination, searching, dan sorting
- **Transaksi Pembelian**: Customer dapat membeli 1 atau lebih produk dengan validasi stok otomatis
- **Authentication & Authorization**:
  - Registrasi customer
  - Login dengan JWT token
  - Role-based access control (Admin & Customer)
- **In-Memory Storage**: Data disimpan di memory dengan default data
- **Swagger Documentation**: API documentation untuk testing

## Requirements

- Go 1.19 atau lebih tinggi
- Swag CLI (untuk regenerate swagger docs)

## Installation

1. Clone atau download project ini

2. Install dependencies:
```bash
cd tokonyadia
go mod download
```

3. (Opsional) Install swag CLI untuk regenerate swagger docs:
```bash
go install github.com/swaggo/swag/cmd/swag@v1.8.12
```

## Running the Application

### Development Mode

```bash
set JWT_SECRET=replace-with-a-random-secret-of-at-least-32-characters
go run cmd/api/main.go
```

### Build untuk Production

Build untuk OS yang sama:
```bash
go build -o tokonyadia-api cmd/api/main.go
./tokonyadia-api
```

Build untuk berbagai OS:

**Linux:**
```bash
GOOS=linux GOARCH=amd64 go build -o tokonyadia-api-linux cmd/api/main.go
```

**Windows:**
```bash
GOOS=windows GOARCH=amd64 go build -o tokonyadia-api.exe cmd/api/main.go
```

**macOS:**
```bash
GOOS=darwin GOARCH=amd64 go build -o tokonyadia-api-macos cmd/api/main.go
```

**macOS ARM (M1/M2):**
```bash
GOOS=darwin GOARCH=arm64 go build -o tokonyadia-api-macos-arm64 cmd/api/main.go
```

## API Documentation

Setelah aplikasi berjalan, akses Swagger documentation di:
```
http://localhost:8080/swagger/index.html
```

## Default Accounts

### Admin Account
- Email: `admin@tokonyadia.com`
- Password: `admin123`
- Role: Admin

### Customer Accounts
- Email: `budi@email.com` / Password: `password123`
- Email: `siti@email.com` / Password: `password123`

## Default Products

1. Laptop ASUS ROG - Rp 15,000,000 (Stock: 10)
2. Mouse Logitech MX Master - Rp 1,200,000 (Stock: 25)
3. Keyboard Mechanical - Rp 850,000 (Stock: 15)
4. Monitor LG 27 inch - Rp 3,500,000 (Stock: 8)
5. Headset Gaming - Rp 750,000 (Stock: 20)

## API Endpoints

### Authentication (Public)
- `POST /api/v1/auth/register` - Register customer baru
- `POST /api/v1/auth/login` - Login dan mendapatkan JWT token

### Products
- `GET /api/v1/products` - Get all products (Public - tanpa auth)
- `GET /api/v1/products/:id` - Get product by ID (Public - tanpa auth)
- `POST /api/v1/products` - Create product (Admin only)
- `PUT /api/v1/products/:id` - Update product (Admin only)
- `DELETE /api/v1/products/:id` - Delete product (Admin only)

### Customers (Admin Only)
- `GET /api/v1/customers` - Get all customers
- `GET /api/v1/customers/:id` - Get customer by ID
- `PUT /api/v1/customers/:id` - Update customer
- `DELETE /api/v1/customers/:id` - Delete customer

### Transactions (Authenticated Users)
- `POST /api/v1/transactions` - Create transaction (purchase products)
- `GET /api/v1/transactions` - Get all transactions (Admin: all, Customer: own transactions)
- `GET /api/v1/transactions/me` - Get my transaction history (Customer's own transactions)
- `GET /api/v1/transactions/:id` - Get transaction by ID
- `GET /api/v1/transactions/customer/:customerId` - Get transactions by customer ID (Admin only)

## Query Parameters

Semua endpoint GET list mendukung:
- `page` - Nomor halaman (default: 1)
- `pageSize` - Jumlah data per halaman (default: 10, max: 100)
- `search` - Pencarian berdasarkan nama/email
- `sortBy` - Field untuk sorting (productName, stock, productPrice, fullname, email, birthdate, transactionDate, customerName)
- `sortDir` - Arah sorting: asc atau desc (default: asc)

Contoh:
```
GET /api/v1/products?page=1&pageSize=10&search=laptop&sortBy=productPrice&sortDir=desc
```

## Authentication Header

Untuk endpoint yang memerlukan authentication, sertakan header:
```
Authorization: Bearer <your-jwt-token>
```

## Example Usage

### 1. Register Customer Baru
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "fullname": "John Doe",
    "birthdate": "1995-01-15T00:00:00Z",
    "address": "Jl. Contoh No. 123",
    "phoneNumber": "081234567890",
    "email": "john@email.com",
    "password": "password123"
  }'
```

### 2. Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@email.com",
    "password": "password123"
  }'
```

### 3. Get Products (Tanpa Auth)
```bash
curl http://localhost:8080/api/v1/products
```

### 4. Create Transaction
```bash
curl -X POST http://localhost:8080/api/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -d '{
    "items": [
      {
        "productId": "550e8400-e29b-41d4-a716-446655440001",
        "quantity": 2
      },
      {
        "productId": "550e8400-e29b-41d4-a716-446655440002",
        "quantity": 1
      }
    ]
  }'
```

### 5. Get My Transaction History
```bash
curl -X GET "http://localhost:8080/api/v1/transactions/me?page=1&pageSize=10" \
  -H "Authorization: Bearer <your-token>"
```

### 6. Get Customer Transactions (Admin Only)
```bash
curl -X GET "http://localhost:8080/api/v1/transactions/customer/650e8400-e29b-41d4-a716-446655440002" \
  -H "Authorization: Bearer <admin-token>"
```

## Project Structure

```
tokonyadia/
├── cmd/
│   └── api/
│       └── main.go           # Entry point aplikasi
├── internal/
│   ├── models/               # Data models
│   │   ├── product.go
│   │   ├── customer.go
│   │   ├── transaction.go
│   │   └── common.go
│   ├── repository/           # Data access layer (in-memory storage)
│   │   ├── product_repository.go
│   │   ├── customer_repository.go
│   │   └── transaction_repository.go
│   ├── service/              # Business logic layer
│   │   ├── product_service.go
│   │   ├── customer_service.go
│   │   └── transaction_service.go
│   ├── handlers/             # HTTP handlers
│   │   ├── product_handler.go
│   │   ├── customer_handler.go
│   │   └── transaction_handler.go
│   ├── middleware/           # HTTP middlewares
│   │   └── auth.go
│   └── utils/                # Utility functions
│       └── jwt.go
├── docs/                     # Swagger documentation (generated)
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── go.mod
├── go.sum
└── README.md
```

## Business Rules

1. **Product Management**:
   - Admin dapat CRUD produk
   - Stok tidak boleh negatif
   - Harga produk harus lebih dari 0

2. **Customer Management**:
   - Email harus unik
   - Password minimal 6 karakter
   - Default role adalah "customer"

3. **Transaction**:
   - Customer harus login untuk melakukan transaksi
   - Validasi stok sebelum transaksi
   - Stok berkurang otomatis setelah transaksi berhasil
   - Harga produk saat transaksi disimpan di transaction detail
   - Customer hanya bisa melihat transaksi mereka sendiri
   - Admin dapat melihat semua transaksi

4. **Authorization**:
   - Get products tidak perlu authentication
   - Semua endpoint lain memerlukan authentication
   - Endpoint customers dan management product (POST/PUT/DELETE) hanya untuk admin

## Data Persistence

Data disimpan di memory menggunakan array/slice. Setiap kali aplikasi di-restart, data akan kembali ke default data. Ini sesuai dengan requirement untuk development/testing.

Jika ingin data persisten, implementasi database (PostgreSQL, MySQL, etc) dapat ditambahkan dengan mengubah layer repository.

## Notes

- Set `JWT_SECRET` to a cryptographically random value of at least 32 characters before starting the API. The service will refuse to start without it.
- Default port adalah 8080. Dapat diubah di `cmd/api/main.go`.
- Concurrent access aman karena menggunakan `sync.RWMutex` di repository layer.

## License

MIT License
