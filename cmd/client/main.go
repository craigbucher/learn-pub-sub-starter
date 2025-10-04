package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
    amqp "github.com/rabbitmq/amqp091-go"	// Go AMQP library:
)

func main() {
	fmt.Println("Starting Peril client...")
	// Update the cmd/client package to connect to Rabbit, similar to the cmd/server package:
	const rabbitConnString = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Peril game client connected to RabbitMQ!")

	// Use the ClientWelcome() function in internal/gamelogic to prompt the user for a username:
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("could not get username: %v", err)
	}

	/* use these parameters to call DeclareAndBind:
exchange: peril_direct (this is a constant in the internal/routing package)
queueName: pause.username where username is the user's input. The pause section of the name is the routing key constant in the internal/routing package and is joined by a ..
routingKey: pause (this is a constant in the internal/routing package)
queueType: transient */
	_, queue, err := pubsub.DeclareAndBind(
		conn, 								// conn, established above
		routing.ExchangePerilDirect,		// exchange
		routing.PauseKey + "." + username,	// queueName
		routing.PauseKey,					// key
		pubsub.SimpleQueueTransient,		// queueType
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	// If a signal is received, print a message to the console that the program is shutting down 
	// and close the connection:
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("RabbitMQ connection closed.")
}
