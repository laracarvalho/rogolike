package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/laracarvalho/rogolike/ecs"
)

var position *ecs.Component
var renderable *ecs.Component

func InitializeWorld(startLevel Level) (*ecs.Engine, map[string]ecs.Tag) {
	tags := make(map[string]ecs.Tag)
	engine := ecs.NewEngine()

	startRoom := startLevel.Rooms[0]
	x, y := startRoom.Center()

	player := engine.NewComponent()
	position = engine.NewComponent()
	renderable = engine.NewComponent()
	movable := engine.NewComponent()
	monster := engine.NewComponent()

	playerImg, _, playerErr := ebitenutil.NewImageFromFile("assets/player.png")
	if playerErr != nil {
		log.Fatal(playerErr)
	}

	skellyImg, _, skellyErr := ebitenutil.NewImageFromFile("assets/skelly.png")
	if skellyErr != nil {
		log.Fatal(skellyErr)
	}

	engine.NewEntity().
		AddComponent(player, Player{}).
		AddComponent(renderable, &Renderable{
			Image: playerImg,
		}).
		AddComponent(movable, Movable{}).
		AddComponent(position, &Position{
			X: x,
			Y: y,
		})

	for _, room := range startLevel.Rooms {
		if room.X != startRoom.X {
			mX, mY := room.Center()
			engine.NewEntity().
				AddComponent(monster, Monster{}).
				AddComponent(renderable, &Renderable{
					Image: skellyImg,
				}).
				AddComponent(position, &Position{
					X: mX,
					Y: mY,
				})
		}
	}

	players := ecs.BuildTag(player, position)
	tags["players"] = players

	renderables := ecs.BuildTag(renderable, position)
	tags["renderables"] = renderables

	return engine, tags
}
