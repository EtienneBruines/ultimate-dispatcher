package main

import (
	"engo.io/engo"
	"engo.io/ecs"
)

const (
	title = "Ultimate Dispatcher"
)

type Game struct {}

func (g *Game) Preload() {

}

func (g *Game) Setup(w *ecs.World) {

}

func (g *Game) Type() string {
	return "GameScene"
}

/*
	// Preload is called before loading resources
	Preload()

	// Setup is called before the main loop
	Setup(*ecs.World)

	// Type returns a unique string representation of the Scene, used to identify it
	Type() string
 */

func main() {
	opts := engo.RunOptions{
		Title: title,
		StandardInputs: true,
	}
	engo.Run(opts, Game{})
}