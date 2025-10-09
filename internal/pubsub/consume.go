package pubsub

import (
	"fmt"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Define AckType:
type Acktype int

// Define your SimpleQueueType
type SimpleQueueType int

// Define the constants for your enum:
const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

// declares a named type Acktype:
// iota auto-increments starting at 0 within the block
const (
	Ack Acktype = iota	// 0
	NackDiscard			// 1
	NackRequeue			// 2
)

/* In your internal/pubsub package, create a new function called SubscribeJSON, here's my 
function signature: */
func SubscribeJSON[T any](
    conn *amqp.Connection,
    exchange,
    queueName,
    key string,
    queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	// Update your internal/pubsub.SubscribeJSON function's handler parameter to return an 
	// "acktype" instead of nothing:
    handler func(T) Acktype,
) error {
	// Call DeclareAndBind to make sure that the given queue exists and is bound to the exchange:
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("could not declare and bind queue: %v", err)
	}
	/* Get a new chan of amqp.Delivery structs by using the channel.Consume method.
		- Use an empty string for the consumer name so that it will be auto-generated
		- Set all other parameters to false/nil */
	msgs, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return fmt.Errorf("could not consume messages: %v", err)
	}
	// create the unmarshaller function:
	unmarshaller := func(data []byte) (T, error) {
		// (Could also do these next 2 steps inside the goroutine)
		var target T
		err := json.Unmarshal(data, &target)
		return target, err
	}
	// Start a goroutine that ranges over the channel of deliveries:
	go func() {
		// make sure to close the channel when the goroutine ends:
		defer ch.Close()
		for msg := range msgs {
			// Unmarshal the body (raw bytes) of each message delivery into the (generic) T type:
			target, err := unmarshaller(msg.Body)
			if err != nil {
				fmt.Printf("could not unmarshal message: %v\n", err)
				continue
			}
			// Call the given handler function with the unmarshaled message:
			// (handler is passed in as a function parameter)
			// For testing/debugging purposes, add a log statement alongside each Ack/Nack call 
			// to indicate which action occurred
			switch handler(target) {
			// Ack: msg.Ack(false):
			// Processed successfully
			case Ack:
				msg.Ack(false)
			// NackDiscard: msg.Nack(false, false):
			// Not processed successfully, and should be discarded (to a dead-letter queue 
			// if configured or just deleted entirely)
			case NackDiscard:
				msg.Nack(false, false)
			// msg.Nack(false, true):
			// Not processed successfully, but should be requeued on the same queue to be 
			// processed again (retry)
			case NackRequeue:
				msg.Nack(false, true)
			}
		}
	}()
	// return nil, since there was no error:
	return nil
}

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
		// RabbitMQ routes any rejected/expired messages from this queue to the specified 
		// dead-letter exchange, which is bound to your dead-letter queue:
		amqp.Table{
			"x-dead-letter-exchange": "peril_dlx",	// Dead-letter exchange
		}, 		
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