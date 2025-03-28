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
	// Connect to RabbitMQ server
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// Create a channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to Open Channel")
	defer ch.Close()

	// Declare a queue
	q, err := ch.QueueDeclare(
		"hello", // queue name
		false,   // durable
		false,   // auto-delete
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)

	failOnError(err, "Failed to declare a Queue")

	// Publish a message
	msg := "Hello, RabbitMQ!"
	publish(ch, q.Name, msg)
}

func publish(ch *amqp.Channel, qName, msg string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := ch.PublishWithContext(ctx, "", qName, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(msg),
	})

	failOnError(err, "Failed to Publish")

	log.Printf(" [x] Sent %s", msg)
}
