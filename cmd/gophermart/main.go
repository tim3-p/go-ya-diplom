package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	"github.com/tim3-p/go-ya-diplom/config"
	"github.com/tim3-p/go-ya-diplom/internal/handlers"
	"github.com/tim3-p/go-ya-diplom/internal/interfaces"
	middleware "github.com/tim3-p/go-ya-diplom/internal/middlewares"
	"github.com/tim3-p/go-ya-diplom/internal/service"
	"github.com/tim3-p/go-ya-diplom/internal/storage"

	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	cfg := config.InitConfig()
	db, err := sql.Open("pgx", cfg.DatabasURI)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("could not start sql migration... %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", cfg.MigrationDir), "product", driver)
	if err != nil {
		log.Fatalf("migration failed... %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("An error occurred while syncing the database.. %v", err)
	}

	userRepository := storage.CreateUser(db)
	orderRepository := storage.CreateOrder(db)
	withdrawalRepository := storage.CreateWithdrawal(db)
	cookieAuthenticator := service.NewCookieAuthenticator([]byte(cfg.Key))
	accrualService := service.NewAccrual(cfg.AccrualSystemAddress, orderRepository)
	accrualService.Start()
	authenticator := middleware.NewAuthenticator(cookieAuthenticator)

	mws := []interfaces.Middleware{
		middleware.GzipEncoder{},
		middleware.GzipDecoder{},
	}

	handler := handlers.NewHandler(
		cfg.AccrualSystemAddress,
		userRepository,
		orderRepository,
		withdrawalRepository,
		cookieAuthenticator,
		accrualService,
		authenticator,
		mws,
	)
	server := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: handler,
	}

	log.Fatal(server.ListenAndServe())
}
