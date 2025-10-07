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
// - takes a pointer to a gamelogic.GameState and returns another function
// - The returned function has the signature func(routing.PlayingState)
// - Because it closes over gs, the inner function can use that game state when a routing.PlayingState 
//   message arrives
//	 (In other words, the returned function forms a closure: its environment includes gs, so you don’t 
//   need to pass gs each time)
func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		// defer a print statement that gives the user a new prompt: defer fmt.Print("> "):
		defer fmt.Print("> ")
		// Use the game state's HandlePause method to pause the game for the client:
		gs.HandlePause(ps)
	}
}

// The handler for new messages should use the GameState's HandleMove method and then print 
// a new > prompt for the user:
// (explanations above)
func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) {
	return func(move gamelogic.ArmyMove) {
		defer fmt.Print("> ")
		gs.HandleMove(move)
	}
}