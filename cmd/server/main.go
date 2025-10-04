package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
    amqp "github.com/rabbitmq/amqp091-go"	// Go AMQP library:
)

func main() {
	fmt.Println("Starting Peril server...")
	// Declare a connection string (This is how your application will know where to connect 
	// to the RabbitMQ server)
	const rabbitConnString = "amqp://guest:guest@localhost:5672/"
	// Call amqp.Dial with the connection string to create a new connection to RabbitMQ:
	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	// Defer a .Close() of the connection to ensure it's closed when the program exits:
	defer conn.Close()
	// Print a message to the console that the connection was successful:
	fmt.Println("Peril game server connected to RabbitMQ!")
	
	// create a new channel using the .Channel method on the connection:
	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	// use the PublishJSON function to publish a message to the exchange:
	// PublishJSON from internal/pubsub/publish.go
	err = pubsub.PublishJSON(
		// Use the channel you created:
		publishCh,
		// Use the internal/routing package's ExchangePerilDirect string for the exchange:
		routing.ExchangePerilDirect,
		// Use the internal/routing package's PauseKey string for the routing key:
		routing.PauseKey,
		// The data to send is the JSON-marshaled internal/routing's PlayingState struct, 
		// with the IsPaused field set to true:
		routing.PlayingState{
			IsPaused: true,
		},
	)
	if err != nil {
		log.Printf("could not publish time: %v", err)
	}
	fmt.Println("Pause message sent!")

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	// If a signal is received, print a message to the console that the program is shutting down 
	// and close the connection:
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("RabbitMQ connection closed.")
	
}
