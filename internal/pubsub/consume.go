package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Define your SimpleQueueType
type SimpleQueueType int

// Define the constants for your enum:
const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

	// Declare and bind a transient queue by creating and using a new function in the 
	// internal/pubsub package:
	func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error){
	// Create a new .Channel() on the connection:
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not create channel: %v", err)
	}
	//Declare a new queue using .QueueDeclare():
	queue, err := ch.QueueDeclare(
		queueName,	// name
		queueType == SimpleQueueDurable, 	// The durable parameter should only be true if queueType is durable
		queueType != SimpleQueueDurable,	// The autoDelete parameter should be true if queueType is transient
		queueType != SimpleQueueDurable,	// The exclusive parameter should be true if queueType is transient
		false,		// The noWait parameter should be false
		nil, 		// The args parameter should be nil
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not declare queue: %v", err)
	}
	// Bind the queue to the exchange using .QueueBind():
	err = ch.QueueBind(queue.Name, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not bind queue: %v", err)
	}
	// Return the channel and queue
	return ch, queue, nil
}