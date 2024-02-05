package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/laracarvalho/rogolike/ecs"
)

var position *ecs.Component
var renderable *ecs.Component
var monster *ecs.Component
var health *ecs.Component
var meleeWeapon *ecs.Component
var armor *ecs.Component
var name *ecs.Component
var userMessage *ecs.Component

func InitializeWorld(startLevel Level) (*ecs.Engine, map[string]ecs.Tag) {
	tags := make(map[string]ecs.Tag)
	engine := ecs.NewEngine()

	player := engine.NewComponent()
	position = engine.NewComponent()
	renderable = engine.NewComponent()
	movable := engine.NewComponent()
	monster = engine.NewComponent()
	health = engine.NewComponent()
	meleeWeapon = engine.NewComponent()
	armor = engine.NewComponent()
	name = engine.NewComponent()
	userMessage = engine.NewComponent()

	startRoom := startLevel.Rooms[0]
	x, y := startRoom.Center()

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
		}).
		AddComponent(health, &Health{
			MaxHealth:     30,
			CurrentHealth: 30,
		}).
		AddComponent(meleeWeapon, &MeleeWeapon{
			Name:          "Fist",
			MinimumDamage: 1,
			MaximumDamage: 3,
			ToHitBonus:    2,
		}).
		AddComponent(armor, &Armor{
			Name:       "Burlap Sack",
			Defense:    1,
			ArmorClass: 1,
		}).
		AddComponent(name, &Name{Label: "Player"}).
		AddComponent(userMessage, &UserMessage{
			AttackMessage:    "",
			DeadMessage:      "",
			GameStateMessage: "",
		})

	for _, room := range startLevel.Rooms {
		if room.X != startRoom.X {
			mX, mY := room.Center()
			engine.NewEntity().
				AddComponent(monster, &Monster{}).
				AddComponent(renderable, &Renderable{
					Image: skellyImg,
				}).
				AddComponent(position, &Position{
					X: mX,
					Y: mY,
				}).
				AddComponent(health, &Health{
					MaxHealth:     10,
					CurrentHealth: 10,
				}).
				AddComponent(meleeWeapon, &MeleeWeapon{
					Name:          "Short Sword",
					MinimumDamage: 1,
					MaximumDamage: 4,
					ToHitBonus:    0,
				}).
				AddComponent(armor, &Armor{
					Name:       "Bone",
					Defense:    3,
					ArmorClass: 4,
				}).
				AddComponent(name, &Name{Label: "Skeleton"}).
				AddComponent(userMessage, &UserMessage{
					AttackMessage:    "",
					DeadMessage:      "",
					GameStateMessage: "",
				})
		}
	}

	players := ecs.BuildTag(player, position, health, meleeWeapon, armor, name, userMessage)
	tags["players"] = players

	renderables := ecs.BuildTag(renderable, position)
	tags["renderables"] = renderables

	monsters := ecs.BuildTag(monster, position, health, meleeWeapon, armor, name, userMessage)
	tags["monsters"] = monsters

	messengers := ecs.BuildTag(userMessage)
	tags["messengers"] = messengers

	return engine, tags
}
