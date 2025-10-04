package pubsub

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Create an exported PublishJSON function in the internal/pubsub package. Here's its signature:
func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	// Marshal the val to JSON bytes:
	dat, err := json.Marshal(val)
	if err != nil {
		return err
	}
	// Use the channel's .PublishWithContext method to publish the message to the exchange 
	// with the routing key:
		// Set ctx to context.Background()
		// Set mandatory to false
		// Set immediate to false.
		// In the amqp.Publishing struct, you only need to set two fields:
			// ContentType to "application/json"
			// Body to the JSON bytes
	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        dat,
	})
}
