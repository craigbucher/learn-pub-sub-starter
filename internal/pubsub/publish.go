package pubsub

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

/* Create an exported PublishJSON function in the internal/pubsub package. Here's its signature:
	func PublishJSON[T any]: Uses generics. T is a type parameter; any means callers can pass any type for val
		The caller supplies T implicitly from the argument, e.g., val
	ch *amqp.Channel: A pointer to a RabbitMQ channel used to publish messages
	exchange, key string: The exchange name and routing key to publish to
	val T: The value to publish; its type is whatever T is (struct, map, etc.)
	error: Returns an error if publishing fails */
func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	// Marshal the val to JSON bytes:
	dat, err := json.Marshal(val)
	if err != nil {
		return err
	}
	/* Use the channel's .PublishWithContext method to publish the message to the exchange 
	with the routing key:
	(PublishWithContext is defined in amqp091-go package. It’s a method on type *amqp.Channel)
		Set ctx to context.Background()
			context.Background() returns an empty, non-cancelable root context. It’s typically 
			used as the top-level context when you don’t have a request-scoped or timeout/cancel 
			context to pass down.
			(In Go, a context carries deadlines, cancelation signals, and request-scoped values 
			across API boundaries.)
		Set mandatory to false
		Set immediate to false.
		In the amqp.Publishing struct, you only need to set two fields:
			ContentType to "application/json"
			Body to the JSON bytes */
	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        dat,
	})
}

