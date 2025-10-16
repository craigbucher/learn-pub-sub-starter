package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

/* That signature declares a function named 'handlerLogs' that returns another function. 
Specifically:
	- handlerLogs takes no parameters
	- It returns a function with this type: func(gamelog routing.GameLog) pubsub.Acktype
So it’s a higher-order function that produces a handler. The returned handler:
	- Accepts a routing.GameLog (your decoded message).
	- Returns a pubsub.Acktype (e.g., Ack/Nack) to tell the consumer how to acknowledge the message
Typical use: you call handlerLogs() to get the handler function, then pass that handler into your 
subscribe function so each incoming message is processed and acknowledged appropriately */
	func handlerLogs() func(gamelog routing.GameLog) pubsub.Acktype {
	/* creates a function literal with signature func(gamelog routing.GameLog) pubsub.Acktype.
	Because it’s inside another function, it can capture variables from the outer scope.
	The caller receives this function and can invoke it later with a routing.GameLog, and it must 
	return a pubsub.Acktype */
	return func(gamelog routing.GameLog) pubsub.Acktype {
		// Defer printing a new prompt to the console:
		defer fmt.Print("> ")
		// Use the gamelogic.WriteLog function to write the log to disk:
		err := gamelogic.WriteLog(gamelog)
		if err != nil {
			fmt.Printf("error writing log: %v\n", err)
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}