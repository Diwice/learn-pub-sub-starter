package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(rps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(rps)
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) {
	return func(gam gamelogic.ArmyMove) {
		defer fmt.Print("> ")
		gs.HandleMove(gam)
	}
}
