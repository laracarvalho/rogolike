package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

var hudErr error = nil

func ProcessHUD(g *Game, screen *ebiten.Image) {
	gd := NewGameData()

	uiY := (gd.ScreenHeight - gd.UIHeight) * gd.TileHeight
	uiX := (gd.ScreenWidth * gd.TileWidth) / 2
	var fontX = uiX + 16
	var fontY = uiY + 24

	for _, p := range g.World.Query(g.WorldTags["players"]) {
		h := p.Components[health].(*Health)
		healthText := fmt.Sprintf("Health: %d / %d", h.CurrentHealth, h.MaxHealth)
		text.Draw(screen, healthText, mplusNormalFont, fontX, fontY, color.White)
		fontY += 16
		ac := p.Components[armor].(*Armor)
		acText := fmt.Sprintf("Armor Class: %d", ac.ArmorClass)
		text.Draw(screen, acText, mplusNormalFont, fontX, fontY, color.White)
		fontY += 16
		defText := fmt.Sprintf("Defense: %d", ac.Defense)
		text.Draw(screen, defText, mplusNormalFont, fontX, fontY, color.White)
		fontY += 16
		wpn := p.Components[meleeWeapon].(*MeleeWeapon)
		dmg := fmt.Sprintf("Damage: %d - %d", wpn.MinimumDamage, wpn.MaximumDamage)
		text.Draw(screen, dmg, mplusNormalFont, fontX, fontY, color.White)
		fontY += 16
		bonus := fmt.Sprintf("To Hit Bonus: %d", wpn.ToHitBonus)
		text.Draw(screen, bonus, mplusNormalFont, fontX, fontY, color.White)
	}
}
