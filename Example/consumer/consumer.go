package main

import (
	"log"

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
	failOnError(err, "Failed to connect RabbitMQ...")
	defer conn.Close()

	// Create a channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to create channel")
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

	failOnError(err, "Failed to declare a queue")

	// Consume messages from the queue
	msgs, err := ch.Consume(
		q.Name, // queue name
		"",     // consumer tag
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // arguments
	)

	failOnError(err, "Failed to register a consumer")

	// Create a channel to wait for messages
	forever := make(chan bool)

	// Start a goroutine to process messages
	go func() {
		for msg := range msgs {
			log.Printf(" [x] Received %s", msg.Body)
		}
	}()

	log.Println(" [*] Waiting for messages. To exit, press CTRL+C")
	<-forever
}
