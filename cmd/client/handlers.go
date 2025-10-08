package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
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
// Update your client's "move" and "pause" handlers to return an "acktype":
func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {
	return func(ps routing.PlayingState) pubsub.Acktype {
		// defer a print statement that gives the user a new prompt: defer fmt.Print("> "):
		defer fmt.Print("> ")
		// Use the game state's HandlePause method to pause the game for the client:
		gs.HandlePause(ps)
		// The "pause" handler should always Ack:
		return pubsub.Ack
	}
}

// The handler for new messages should use the GameState's HandleMove method and then print 
// a new > prompt for the user:
// (explanations above)
// Update your client's "move" and "pause" handlers to return an "acktype":
func handlerMove(gs *gamelogic.GameState, publishCh *amqp.Channel) func(gamelogic.ArmyMove) pubsub.Acktype {
// func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(move gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(move)
		switch moveOutcome {
		// The "move" handler should "NackDiscard" if:
			// The move outcome was "same player":
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.Ack
		// The "move" handler should only "Ack" if:
			// The move outcome was "safe":
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
			// Update the "move" handler so that when detects MoveOutcomeMakeWar:
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilTopic,
				// Publish a message to the "topic" exchange with the routing key:
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),	// $WARPREFIX.$USERNAME
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				fmt.Printf("error: %s\n", err)
				return pubsub.NackRequeue
			}
			// NackRequeue the message: (Might seem crazy, but it will be fun)
			return pubsub.NackRequeue
		}
		//  "NackDiscard" if the move outcome was anything else:
		fmt.Println("error: unknown move outcome")
		return pubsub.NackDiscard
	}
}

// Create a new handler that consumes all the war messages that the "move" handler publishes, 
// no matter the username in the routing key. It should:
func handlerWar(gs *gamelogic.GameState) func(dw gamelogic.RecognitionOfWar) pubsub.Acktype {
	return func(dw gamelogic.RecognitionOfWar) pubsub.Acktype {
		// defer fmt.Print("> ") to ensure a new prompt is printed after the handler is done:
		defer fmt.Print("> ")
		// Call the gamestate's HandleWar method with the message's body:
		warOutcome, _, _ := gs.HandleWar(dw)
		switch warOutcome {
		// If the outcome is gamelogic.WarOutcomeNotInvolved: NackRequeue the message so another 
		// client can try to consume it:
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		// If the outcome is gamelogic.WarOutcomeNoUnits: NackDiscard the message:
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		// If the outcome is gamelogic.WarOutcomeOpponentWon: Ack the message:
		case gamelogic.WarOutcomeOpponentWon:
			return pubsub.Ack
		// If the outcome is gamelogic.WarOutcomeYouWon: Ack the message:
		case gamelogic.WarOutcomeYouWon:
			return pubsub.Ack
		// If the outcome is gamelogic.WarOutcomeDraw: Ack the message:
		case gamelogic.WarOutcomeDraw:
			return pubsub.Ack
		}
		// If it's anything else, print an error and NackDiscard the message:
		fmt.Println("error: unknown war outcome")
		return pubsub.NackDiscard
	}
}