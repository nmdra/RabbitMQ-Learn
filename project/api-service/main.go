package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Notification struct {
	ID      uint `json:"id"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func main() {
	http.HandleFunc("/send-notification", sendNotificationHandler)

	slog.Info("API Gateway is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func sendNotificationHandler(w http.ResponseWriter, r *http.Request) {
	var notification Notification
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&notification)
	if err != nil {
		http.Error(w, "Failed to parse request", http.StatusBadRequest)
		return
	}

	// Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq-notification:5672")
	if err != nil {
		http.Error(w, "Failed to connect to RabbitMQ", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		http.Error(w, "Failed to open a channel", http.StatusInternalServerError)
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
		http.Error(w, "Failed to declare exchange", http.StatusInternalServerError)
		return
	}

	// Declare a queue with TTL (message TTL will be applied here)
	queue, err := ch.QueueDeclare(
		"sms_notifications", // Queue name
		true,                // Durable
		false,               // Auto-delete
		false,               // Exclusive
		false,               // No-wait
		nil,
		// map[string]interface{}{
		// 	"x-message-ttl": int32(60000), // Set TTL to 60,000 ms (1 minute)
		// },
	)
	if err != nil {
		http.Error(w, "Failed to declare queue", http.StatusInternalServerError)
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
		http.Error(w, "Failed to bind queue", http.StatusInternalServerError)
		return
	}

	// Serialize the notification object to JSON
	body, err := json.Marshal(notification)
	if err != nil {
		http.Error(w, "Failed to serialize notification", http.StatusInternalServerError)
		return
	}

	// Publish the notification to the fanout exchange
	err = ch.Publish(
		"notifications_exchange", // Exchange name
		"",                       // Routing key (not needed for fanout)
		true,                     // Mandatory
		false,                    // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body, // Use the serialized notification JSON
			Expiration:  "600000", // TTL set to 60000 ms (1 minute)
		},
	)
	if err != nil {
		http.Error(w, "Failed to publish a message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Notification triggered"))
}
