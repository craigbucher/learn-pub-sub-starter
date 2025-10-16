package main

import (
	"fmt"
	"log"
	// "os"
	// "os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
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
	defer publishCh.Close()

	/* Update the cmd/server application to declare and bind a queue to the new peril_topic exchange.
		- It should be a durable queue named game_logs.
		- The routing key should be game_logs.*. We'll go into detail on the routing key later. */
	/* Update the server to SubscribeGob to the game_logs queue instead of just declaring it. 
	Use a wildcard in the routing key to make sure you capture logs from all clients, no matter 
	the username */
	err = pubsub.SubscribeGob(
		conn, 								// conn, established above
		routing.ExchangePerilTopic,			// exchange
		routing.GameLogSlug,				// queueName
		routing.GameLogSlug+".*",			// key
		pubsub.SimpleQueueDurable,			// queueType
		handlerLogs(),
	)
	if err != nil {
		log.Fatalf("could not starting consuming logs: %v", err)
	}

	// Run the PrintServerHelp function in internal/gamelogic as the server starts up so that 
	// you can see the commands the user of the REPL can use:
	gamelogic.PrintServerHelp()

	// start an infinite loop:
	for {
		// At the beginning of the loop, use the GetInput function in internal/gamelogic to wait 
		// for a slice of input "words" from the user. If the slice is empty, continue to the 
		// next iteration of the loop:
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		// Check the first word:
		switch input[0] {
			// If it's "pause", log to the console that you're sending a pause message, and publish 
			// the pause message as you were doing before
			case "pause":
				fmt.Println("sending a pause message")
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
					log.Printf("could not publish message: %v", err)
				}
				fmt.Println("Pause message sent!")
			// If it's "resume", log to the console that you're sending a resume message, and publish 
			// the resume message as you were doing before. The only difference is that the IsPaused 
			// field should be set to false:
			case "resume":
				fmt.Println("sending a resume message")
				err = pubsub.PublishJSON(
					publishCh,
					routing.ExchangePerilDirect,
					routing.PauseKey,
					routing.PlayingState{
						IsPaused: false,
					},
				)
				if err != nil {
					log.Printf("could not publish message: %v", err)
				}
				fmt.Println("Resume message sent!")
			// If it's "quit", log to the console that you're exiting, and break out of the loop:
			case "quit":
				fmt.Println("Quitting. . .")
				return
			// If it's anything else, log to the console that you don't understand the command:
			default:
				fmt.Println("I don't understand that command")
		}
	}

	// // wait for ctrl+c
	// signalChan := make(chan os.Signal, 1)
	// // If a signal is received, print a message to the console that the program is shutting down 
	// // and close the connection:
	// signal.Notify(signalChan, os.Interrupt)
	// <-signalChan
	// fmt.Println("RabbitMQ connection closed.")
	
}
