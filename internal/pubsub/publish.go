package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
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

// Add a PublishGob function to the internal/pubsub package
// It should be similar to the PublishJSON function, but encode to gob:
	/* [T any] makes this a generic function. The T is a type parameter that can be any type. 
	This allows you to call PublishGob with different types without rewriting the function for 
	each one */
	/* ch *amqp.Channel - A pointer to an AMQP channel, which is your connection to RabbitMQ 
	for publishing messages */
	/* val T - The value to be published. Its type is T, which means it can be any type you 
	specify when calling the function */
func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	// Create a bytes buffer to collect the encoded data:
	var buffer bytes.Buffer
	// Build a gob encoder that writes into that buffer:
	encoder := gob.NewEncoder(&buffer)
	// Serialize val into gob format:
	err := encoder.Encode(val)
	if err != nil {
		return err
	}
	// publish a message to RabbitMQ through the AMQP channel:
		/* context.Background() - A context for the operation. Using Background() means there's 
		no timeout or cancellation set up (it's a basic, non-cancelable context) */
		/* false (first one) - The mandatory flag. When false, if the message can't be routed 
		to any queue, it's silently dropped. If true, RabbitMQ would return an error */
		/* false (second one) - The immediate flag. When false, the message can wait in a queue. 
		If true, it would require an immediate consumer (this is deprecated in modern RabbitMQ) */
		// amqp.Publishing{...} - The actual message being published:
	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",		// Set the ContentType option to application/gob
		Body:        buffer.Bytes(), 	// The actual message content as a byte slice (the gob-encoded data from your buffer)
	})
}