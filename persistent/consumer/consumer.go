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
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Declare the durable exchange
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

	// Declare a **named** and **durable** queue
	q, err := ch.QueueDeclare(
		"persistent_logs_queue", // Queue name (named to persist messages)
		true,                    // Durable (messages survive broker restart)
		false,                   // Delete when unused
		false,                   // Exclusive
		false,                   // No-wait
		nil,                     // Arguments
	)
	failOnError(err, "Failed to declare a queue")

	// Bind the queue to the fanout exchange
	err = ch.QueueBind(
		q.Name, // Queue name
		"",     // Routing key (ignored in fanout)
		"logs", // Exchange name
		false,  // No-wait
		nil,    // Arguments
	)
	failOnError(err, "Failed to bind queue")

	// Start consuming messages
	msgs, err := ch.Consume(
		q.Name, // Queue name
		"",     // Consumer name
		true,   // Auto-ack
		false,  // Exclusive
		false,  // No-local
		false,  // No-wait
		nil,    // Arguments
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan struct{})
	go func() {
		for d := range msgs {
			log.Printf(" [x] Received: %s", d.Body)
		}
	}()

	log.Println(" [*] Waiting for messages. To exit, press CTRL+C")
	<-forever
}
