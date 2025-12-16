package queue

import (
	"context"
	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	Queue   amqp.Queue
}

// Connect to RabbitMQ and declare the queue
func ConnectRabbitMQ(url string) *RabbitMQ {
	// 1. Connect
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatal("❌ Failed to connect to RabbitMQ:", err)
	}

	// 2. Open Channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("❌ Failed to open a channel:", err)
	}

	// 3. Declare Queue (Must ensure it exists)
	q, err := ch.QueueDeclare(
		"chat_messages", // Queue name
		true,            // Durable (messages survive restart)
		false,           // Delete when unused
		false,           // Exclusive
		false,           // No-wait
		nil,             // Arguments
	)
	if err != nil {
		log.Fatal("❌ Failed to declare a queue:", err)
	}

	log.Println("✅ Connected to RabbitMQ")
	return &RabbitMQ{Conn: conn, Channel: ch, Queue: q}
}

// Publish sends a message to the queue
func (r *RabbitMQ) Publish(ctx context.Context, body interface{}) error {
	bytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	return r.Channel.PublishWithContext(ctx,
		"",           // Exchange
		r.Queue.Name, // Routing key
		false,        // Mandatory
		false,        // Immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent, // Persist to disk
			ContentType:  "application/json",
			Body:         bytes,
			Timestamp:    time.Now(),
		},
	)
}

func (r *RabbitMQ) Close() {
	r.Channel.Close()
	r.Conn.Close()
}
