package main

import (
	"context"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Declare a fanout exchange
	err = ch.ExchangeDeclare(
		"logs",   // Exchange name
		"fanout", // Exchange type
		true,     // Durable
		false,    // Auto-delete
		false,    // Internal
		false,    // No-wait
		nil,      // Arguments
	)
	failOnError(err, "Failed to declare an exchange")

	// Publish a message
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := "Hello, Multiple Consumers!"
	err = ch.PublishWithContext(
		ctx,
		"logs", // Exchange name
		"",     // Routing key (ignored in fanout)
		false,  // Mandatory
		false,  // Immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")

	log.Printf(" [x] Sent %s", body)
}
