package main

import (
	"log"
	"net/http"
	"os"

	httpdelivery "kitchen/internal/delivery/http"
	"kitchen/internal/kafka"
	"kitchen/internal/repository"
	"kitchen/internal/services"
)

func main() {
	repo := repository.NewConfiguredRepository()
	publisher := kafka.NewLoggingPublisher(log.Default())
	catalog := services.NewCatalogService(repo)
	orders := services.NewOrderService(repo, publisher)
	handler := httpdelivery.NewRouter(catalog, orders)
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("kitchen-api listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
