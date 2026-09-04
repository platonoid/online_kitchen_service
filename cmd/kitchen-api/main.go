package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	httpdelivery "kitchen/internal/delivery/http"
	"kitchen/internal/kafka"
	"kitchen/internal/repository"
	"kitchen/internal/services"
)

func main() {
	repo := repository.NewConfiguredRepository()
	brokers := os.Getenv("KAFKA_BROKERS")
	topic := os.Getenv("KAFKA_ORDERS_TOPIC")
	if topic == "" {
		topic = "orders.events"
	}
	var publisher services.EventPublisher = kafka.NewLoggingPublisher(log.Default())
	var consumer *kafka.Consumer
	if brokers != "" {
		publisher = kafka.NewKafkaPublisher(brokers, topic)
		consumer = kafka.NewConsumer(brokers, topic, "kitchen-api")
	}
	catalog := services.NewCatalogService(repo)
	orders := services.NewOrderService(repo, publisher)
	if consumer != nil {
		go func() {
			if err := consumer.Run(context.Background(), orders.HandleRestaurantEvent); err != nil && !strings.Contains(err.Error(), "context canceled") {
				log.Printf("kafka consumer stopped: %v", err)
			}
		}()
	}
	handler := httpdelivery.NewRouter(catalog, orders)
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("kitchen-api listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
