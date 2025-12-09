package app

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/romanWienicke/go-app-test/foundation/postgres"
	"github.com/romanWienicke/go-app-test/rest"
	customerService "github.com/romanWienicke/go-app-test/service/customer"
	orderService "github.com/romanWienicke/go-app-test/service/order"
	productService "github.com/romanWienicke/go-app-test/service/product"
	userService "github.com/romanWienicke/go-app-test/service/user"
)

type app struct {
	db              *postgres.Db
	userService     *userService.UserService
	orderService    *orderService.OrderService
	customerService *customerService.CustomerService
	productService  *productService.ProductService
}

func NewApp() (*app, error) {
	app := &app{}
	if err := app.init(); err != nil {
		return nil, err
	}

	app.userService = userService.NewUserService(app.db)
	app.orderService = orderService.NewOrderService(app.db)
	app.customerService = customerService.NewCustomerService(app.db)
	app.productService = productService.NewProductService(app.db)

	return app, nil
}

func (a *app) Run() {
	a.runServer()
}

func (a *app) runServer() {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}
	// Start the REST server in a goroutine

	go func() {
		if err := rest.NewServer(port,
			a.userService.RouteAdder(),
			a.orderService.RouteAdder(),
			a.customerService.RouteAdder(),
			a.productService.RouteAdder()); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Create a channel to listen for interrupt or terminate signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gracefully...")
	if err := a.Close(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}

func (a *app) Close() error {
	if a.db == nil {
		return nil
	}
	return a.db.Close()
}

func (a *app) init() error {
	var errs error

	errs = errors.Join(errs, a.initPostgres())

	a.userService = userService.NewUserService(a.db)
	a.orderService = orderService.NewOrderService(a.db)
	a.customerService = customerService.NewCustomerService(a.db)
	a.productService = productService.NewProductService(a.db)
	return errs
}

func (a *app) initPostgres() error {
	dbConfig := postgres.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
	}
	var err error
	a.db, err = postgres.NewPostgres(dbConfig)
	if err != nil {
		return err
	}

	return a.db.Init("../migrations")
}
