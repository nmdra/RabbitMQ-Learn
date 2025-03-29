package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/nmdra/RabbitMQ-Learn/project/sms-service/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq-notification:5672/")
	if err != nil {
		slog.Error("Failed to connect to RabbitMQ", slog.Any("error", err))
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		slog.Error("Failed to open channel", slog.Any("error", err))
		return
	}
	defer ch.Close()

	// Declare a fanout exchange
	err = ch.ExchangeDeclare(
		"notifications_exchange", // Exchange name
		"fanout",                 // Exchange type
		true,                     // Durable
		false,                    // Auto-delete
		false,                    // Internal
		false,                    // No-wait
		nil,                      // Arguments
	)
	if err != nil {
		slog.Error("Failed to declare exchange", slog.Any("error", err))
		return
	}

	// Declare the queue without DLX
	queue, err := ch.QueueDeclare(
		"sms_notifications",
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // no extra args
	)
	if err != nil {
		slog.Error("Failed to declare queue", slog.Any("error", err))
		return
	}

	// Bind the queue to the fanout exchange
	err = ch.QueueBind(
		queue.Name,                // Queue name
		"",                         // Routing key (not needed for fanout)
		"notifications_exchange",   // Exchange name
		false,                      // No-wait
		nil,                         // Arguments
	)
	if err != nil {
		slog.Error("Failed to bind queue", slog.Any("error", err))
		return
	}

	// Register consumer to receive messages from the queue
	msgs, err := ch.Consume(
		queue.Name,
		"",
		false, // Don't auto-acknowledge
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		slog.Error("Failed to register consumer", slog.Any("error", err))
		return
	}

	// Process incoming messages
	for d := range msgs {
		var notification models.Notification
		// Decode the message body into a notification model
		if err := json.Unmarshal(d.Body, &notification); err != nil {
			slog.Error("Error decoding message", slog.Any("error", err))
			d.Nack(false, true) // Requeue the message if decoding fails
			continue
		}

		// Simulate processing SMS
		slog.Info(fmt.Sprintf("Sending SMS: %s", notification.Message))
		time.Sleep(5 * time.Second)

		// Acknowledge the message after successful processing
		d.Ack(false)
	}
}

