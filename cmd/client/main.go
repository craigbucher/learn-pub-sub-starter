package main

import (
	"fmt"
	"log"
	"strconv"
	"time"
	// "os"
	// "os/signal""

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

	// Create a channel to *publish* messages (like army moves) or to consume messages from queues:
	// (similar to server/main.go)
	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

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
	// _, queue, err := pubsub.DeclareAndBind(
	// 	conn, 								// conn, established above
	// 	routing.ExchangePerilDirect,		// exchange
	// 	routing.PauseKey + "." + username,	// queueName
	// 	routing.PauseKey,					// key
	// 	pubsub.SimpleQueueTransient,		// queueType
	// )
	// if err != nil {
	// 	log.Fatalf("could not subscribe to pause: %v", err)
	// }
	// fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	// after declaring and binding the pause queue, use the NewGameState function in 
	// internal/gamelogic to create a new game state (and return a pointer to it):
	gs := gamelogic.NewGameState(username)

	// Each game client should subscribe to moves from other players before starting its REPL.
	/* Bind to the army_moves.* routing key:
		- Use army_moves.username as the queue name, where username is the name of the player
		- Use the peril_topic exchange
		- Use a transient queue */
	err = pubsub.SubscribeJSON(
		conn,									// the connection
		routing.ExchangePerilTopic,				// The direct exchange (constant can be found in internal/routing)
		routing.ArmyMovesPrefix+"."+username, 	// A queue named army_moves.username where username is the username of the player
		routing.ArmyMovesPrefix+".*",			// The routing key army_moves.* (constant can be found in internal/routing)
		pubsub.SimpleQueueTransient,			// Transient queue type
		handlerMove(gs, publishCh),				// From client/handlers.go
	)
	if err != nil {
		log.Fatalf("could not subscribe to army moves: %v", err)
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,				// the connection
		routing.WarRecognitionsPrefix,			// The topic exchange (constant can be found in internal/routing)
		routing.WarRecognitionsPrefix+".*",		// The routing routing.WarRecognitionsPrefix (constant can be found in internal/routing)
		pubsub.SimpleQueueDurable,				// Durable queue type
		handlerWar(gs, publishCh),				// From client/handlers.go
	)
	if err != nil {
		log.Fatalf("could not subscribe to war declarations: %v", err)
	}

	// In the cmd/client package's main function, after creating the game state, call 
	// pubsub.SubscribeJSON with the following parameters:
	err = pubsub.SubscribeJSON(
		conn,									// the connection
		routing.ExchangePerilDirect,			// The direct exchange (constant can be found in internal/routing)
		routing.PauseKey+"."+username,			// A queue named pause.username where username is the username of the player
		routing.PauseKey,						// The routing key pause (constant can be found in internal/routing)
		pubsub.SimpleQueueTransient,			// Transient queue type
		handlerPause(gs),						// From client/handlers.go
	)
	if err != nil {
		log.Fatalf("could not subscribe to Pause: %v", err)
	}

	// Add a REPL loop similar to what you did in the cmd/server application:
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		switch input[0] {
			/* The spawn command allows a player to add a new unit to the map under their control. 
			Use the gamestate.CommandSpawn method and pass in the "words" from the GetInput command.
				- Possible unit types are: infantry, cavalry, artillery
				- Possible locations are: americas, europe, africa, asia, antarctica, australia
				- Example usage: spawn europe infantry
				- After spawning a unit, you should see its ID printed to the console. */
			case "spawn":
				err = gs.CommandSpawn(input)
				if err != nil {
					fmt.Println(err)
					continue
				}
			/* The move command allows a player to move their units to a new location. It accepts 
			two arguments: the destination, and the ID of the unit. Call the gamestate.CommandMove 
			method and pass in all with "words" from the GetInput command. If the move is successful,
			print a message indicating that it worked.
				- Example usage: move europe 1 */
			case "move":
				mv, err := gs.CommandMove(input)
				if err != nil {
					fmt.Println(err)
					continue
				}
				// The move command in the REPL should now publish a move:
				err = pubsub.PublishJSON(
				publishCh,
				// Use the peril_topic exchange:							
				routing.ExchangePerilTopic,
				// Publish the move to the army_moves.username routing key, where username is 
				// the name of the player:
				routing.ArmyMovesPrefix+"."+mv.Player.Username,
				mv,
				)
				if err != nil {
					fmt.Printf("error: %s\n", err)
					continue
				}
				// Log a message to the console stating that the move was published successfully:
				fmt.Printf("Moved %v units to %s\n", len(mv.Units), mv.ToLocation)
			/* The status command uses the gamestate.CommandStatus method to print the current 
			status of the player's game state. */
			case "status":
				gs.CommandStatus()
			/* The help command uses the gamelogic.PrintClientHelp function to print a list of 
			available commands. */
			case "help":
				gamelogic.PrintClientHelp()
			// For now, the spam command just prints a message that says "Spamming not allowed yet!"
			case "spam":
				if len(input) < 2 {
				fmt.Println("usage: spam <n>")
				continue
			}
			n, err := strconv.Atoi(input[1])
			if err != nil {
				fmt.Printf("error: %s is not a valid number\n", input[1])
				continue
			}
			for i := 0; i < n; i++ {
				// Use gamelogic.GetMaliciousLog to get a malicious log message:
				msg := gamelogic.GetMaliciousLog()
				/* Publish the log message (a struct) to Rabbit. Use the following parameters:
				Exchange: peril_topic
				Key: game_logs.username, where username is the username of the player */
				err = publishGameLog(publishCh, username, msg)
				if err != nil {
					fmt.Printf("error publishing malicious log: %s\n", err)
				}
			}
				fmt.Printf("Published %v malicious logs\n", n)
			// The quit command uses the gamelogic.PrintQuit function to print a message, then exit the REPL
			case "quit":
				gamelogic.PrintQuit()
				return
			// If any other command is entered, print an error message and continue the loop:
			default:
				fmt.Println("unknown command")
		}
	}
}

// Create a reusable function to publish a GameLog struct:
	// publishCh *amqp.Channel - A pointer to an AMQP channel for publishing to RabbitMQ
	// username - A string representing the player's username (used in the routing key)
	// msg - A string containing the actual log message (e.g., "player1 won a war against player2")
	// Returns an error (or nil if successful)
	/* This function encapsulates all the logic needed to publish a game log. Instead of repeating 
	the same publishing code every time you need to log something, you can just call this function 
	with the username and message */
func publishGameLog(publishCh *amqp.Channel, username, msg string) error {
	return pubsub.PublishGob(
		publishCh,								// the connection
		routing.ExchangePerilTopic,				// The topic exchange
		routing.GameLogSlug+"."+username,		// A queue named game_logs.username where username is the username of the player that initiated the war
		// The GameLog struct should be serialized using the PublishGob function. Fill all the fields in:
		// Creates a new instance of the GameLog struct from the routing package:
		// This struct represents a complete game log entry
		routing.GameLog{
			Username:    username,
			CurrentTime: time.Now(),	// Set the CurrentTime field to the current timestamp 
			Message:     msg,			// Set the Message field to the value of the msg parameter passed into your function
		},
	)
}



	// // wait for ctrl+c
	// signalChan := make(chan os.Signal, 1)
	// // If a signal is received, print a message to the console that the program is shutting down 
	// // and close the connection:
	// signal.Notify(signalChan, os.Interrupt)
	// <-signalChan
	// fmt.Println("RabbitMQ connection closed.")

