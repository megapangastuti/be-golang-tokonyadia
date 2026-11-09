package main

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"

	"tokonyadia/internal/handlers"
	"tokonyadia/internal/middleware"
	"tokonyadia/internal/repository"
	"tokonyadia/internal/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "tokonyadia/docs"
)

// @title Tokonyadia API
// @version 1.0
// @description REST API for Tokonyadia E-commerce Backend
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@tokonyadia.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Validate JWT secret
	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))

	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET must be set to a random value of at least 32 characters")
	}

	// Initialize repositories
	productRepo := repository.NewProductRepository()
	transactionRepo := repository.NewTransactionRepository()

	customerRepo, err := repository.NewCustomerRepository()
	if err != nil {
		log.Fatalf("Failed to initialize customer repository: %v", err)
	}

	// Initialize services
	productService := service.NewProductService(productRepo)
	customerService := service.NewCustomerService(customerRepo)
	transactionService := service.NewTransactionService(
		transactionRepo,
		productRepo,
		customerRepo,
	)

	// Initialize handlers
	productHandler := handlers.NewProductHandler(productService)
	customerHandler := handlers.NewCustomerHandler(customerService)
	transactionHandler := handlers.NewTransactionHandler(transactionService)

	// Setup router
	router := gin.Default()

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", customerHandler.Register)
			auth.POST("/login", customerHandler.Login)
		}

		// Product routes
		products := v1.Group("/products")
		{
			// Public routes
			products.GET("", productHandler.GetAll)
			products.GET("/:id", productHandler.GetByID)

			// Protected routes (admin only)
			adminProducts := products.Group("")
			adminProducts.Use(
				middleware.AuthMiddleware(),
				middleware.AdminMiddleware(),
			)
			{
				adminProducts.POST("", productHandler.Create)
				adminProducts.PUT("/:id", productHandler.Update)
				adminProducts.DELETE("/:id", productHandler.Delete)
			}
		}

		// Customer routes (admin only)
		customers := v1.Group("/customers")
		customers.Use(
			middleware.AuthMiddleware(),
			middleware.AdminMiddleware(),
		)
		{
			customers.GET("", customerHandler.GetAll)
			customers.GET("/:id", customerHandler.GetByID)
			customers.PUT("/:id", customerHandler.Update)
			customers.DELETE("/:id", customerHandler.Delete)
		}

		// Transaction routes (authenticated users)
		transactions := v1.Group("/transactions")
		transactions.Use(middleware.AuthMiddleware())
		{
			transactions.POST("", transactionHandler.Create)
			transactions.GET("", transactionHandler.GetAll)
			transactions.GET("/me", transactionHandler.GetMyTransactions)
			transactions.GET("/:id", transactionHandler.GetByID)

			// Admin only routes
			adminTransactions := transactions.Group("")
			adminTransactions.Use(middleware.AdminMiddleware())
			{
				adminTransactions.GET(
					"/customer/:customerId",
					transactionHandler.GetCustomerTransactions,
				)
			}
		}
	}

	// Start server
	log.Println("Starting Tokonyadia API server on :8080")
	log.Println("Swagger documentation available at: http://localhost:8080/swagger/index.html")
	log.Println("Default admin credentials: admin@tokonyadia.com / admin123")
	log.Println("Default customer credentials: budi@email.com / password123")

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
