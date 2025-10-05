package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

/* Create a new function called handlerPause in the cmd/client application package. It accepts 
a game state struct and returns a new handler function that accepts a routing.PlayingState struct. 
This will be the handler we pass into SubscribeJSON that will be called each time a new message 
is consumed. Here's my signature: */
func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		// defer a print statement that gives the user a new prompt: defer fmt.Print("> "):
		defer fmt.Print("> ")
		// Use the game state's HandlePause method to pause the game for the client:
		gs.HandlePause(ps)
	}
}