package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Level holds the tile information for a complete dungeon level.
type Level struct {
	Tiles []MapTile
	Rooms []Rect
}

type TileType int

const (
	WALL TileType = iota
	FLOOR
)

// MapTile is a single Tile on a given level
type MapTile struct {
	PixelX   int
	PixelY   int
	Blocked  bool
	Image    *ebiten.Image
	TileType TileType
}

// NewLevel creates a new game level in a dungeon.
func NewLevel() Level {
	l := Level{}
	rooms := make([]Rect, 0)
	l.Rooms = rooms
	l.generateLevelTiles()

	return l
}

// GetIndexFromXY gets the index of the map array from a given X,Y TILE coordinate.
// This coordinate is logical tiles, not pixels.
func (level *Level) GetIndexFromXY(x int, y int) int {
	gd := NewGameData()
	return (y * gd.ScreenWidth) + x
}

func (level *Level) createTiles() []MapTile {
	gd := NewGameData()
	tiles := make([]MapTile, gd.ScreenHeight*gd.ScreenWidth)
	index := 0

	wall, _, wallErr := ebitenutil.NewImageFromFile("assets/wall.png")
	if wallErr != nil {
		log.Fatal(wallErr)
	}

	for x := 0; x < gd.ScreenWidth; x++ {
		for y := 0; y < gd.ScreenHeight; y++ {
			index = level.GetIndexFromXY(x, y)

			tile := MapTile{
				PixelX:   x * gd.TileWidth,
				PixelY:   y * gd.TileHeight,
				Blocked:  true,
				Image:    wall,
				TileType: WALL,
			}

			tiles[index] = tile
		}
	}
	return tiles
}

func (level *Level) DrawLevel(screen *ebiten.Image) {
	gd := NewGameData()
	for x := 0; x < gd.ScreenWidth; x++ {
		for y := 0; y < gd.ScreenHeight; y++ {
			tile := level.Tiles[level.GetIndexFromXY(x, y)]
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(tile.PixelX), float64(tile.PixelY))
			screen.DrawImage(tile.Image, op)
		}
	}
}

func (level *Level) createRoom(room Rect) {
	floor, _, floorErr := ebitenutil.NewImageFromFile("assets/floor.png")
	if floorErr != nil {
		log.Fatal(floorErr)
	}

	for y := room.Y + 1; y < room.Height; y++ {
		for x := room.X + 1; x < room.Width; x++ {
			index := level.GetIndexFromXY(x, y)
			level.Tiles[index].Blocked = false
			level.Tiles[index].TileType = FLOOR
			level.Tiles[index].Image = floor
		}
	}
}

func (level *Level) createHorizontalTunnel(x1 int, x2 int, y int) {
	gd := NewGameData()

	floor, _, err := ebitenutil.NewImageFromFile("assets/floor.png")
	if err != nil {
		log.Fatal(err)
	}

	for x := Min(x1, x2); x < Max(x1, x2)+1; x++ {
		index := level.GetIndexFromXY(x, y)
		if index > 0 && index < gd.ScreenWidth*gd.ScreenHeight {
			level.Tiles[index].Blocked = false
			level.Tiles[index].TileType = FLOOR
			level.Tiles[index].Image = floor
		}
	}
}

func (level *Level) createVerticalTunnel(y1 int, y2 int, x int) {
	gd := NewGameData()

	floor, _, err := ebitenutil.NewImageFromFile("assets/floor.png")
	if err != nil {
		log.Fatal(err)
	}

	for y := Min(y1, y2); y < Max(y1, y2)+1; y++ {
		index := level.GetIndexFromXY(x, y)
		if index > 0 && index < gd.ScreenWidth*gd.ScreenHeight {
			level.Tiles[index].Blocked = false
			level.Tiles[index].TileType = FLOOR
			level.Tiles[index].Image = floor
		}
	}
}

func (level *Level) generateLevelTiles() {
	MIN_SIZE := 6
	MAX_SIZE := 10
	MAX_ROOMS := 30
	contains_rooms := false

	gd := NewGameData()
	tiles := level.createTiles()
	level.Tiles = tiles

	for idx := 0; idx < MAX_ROOMS; idx++ {
		w := GetRandomBetween(MIN_SIZE, MAX_SIZE)
		h := GetRandomBetween(MIN_SIZE, MAX_SIZE)
		x := GetRandomBetween(0, gd.ScreenWidth-w-1)
		y := GetRandomBetween(0, gd.ScreenHeight-h-1)
		room := NewRect(x, y, w, h)

		new_room := NewRect(x, y, w, h)
		okToAdd := true

		for _, otherRoom := range level.Rooms {
			if new_room.Intersect(otherRoom) {
				okToAdd = false
				break
			}
		}

		if okToAdd {
			level.createRoom(new_room)

			if contains_rooms {
				newX, newY := new_room.Center()
				prevX, prevY := level.Rooms[len(level.Rooms)-1].Center()

				coinflip := GetDiceRoll(2)

				if coinflip == 2 {
					level.createHorizontalTunnel(prevX, newX, prevY)
					level.createVerticalTunnel(prevY, newY, newX)

				} else {
					level.createHorizontalTunnel(prevX, newX, newY)
					level.createVerticalTunnel(prevY, newY, prevX)
				}
			}

			level.Rooms = append(level.Rooms, room)
			contains_rooms = true
		}
	}
}
