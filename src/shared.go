package main

import (
	"github.com/hajimehoshi/ebiten"
	"image"
)

// Hitbox is used on many items in the game to determine collisions
type Hitbox struct {
	x float64
	y float64
	w float64
	h float64
}

func collide(x1 float64, y1 float64, w1 float64, h1 float64, x2 float64, y2 float64, w2 float64, h2 float64) bool {
	// Check x and y for overlap
	if x2 > w1+x1 || x1 > w2+x2 || y2 > h1+y1 || y1 > h2+y2 {
		return false
	}
	return true
}

func spriteDraw(screen *ebiten.Image, g *Game, sprite string) {
	screen.DrawImage(spriteAtlas.SubImage(image.Rect(g.sprites[sprite].x, g.sprites[sprite].y, g.sprites[sprite].x+g.sprites[sprite].width, g.sprites[sprite].y+g.sprites[sprite].height)).(*ebiten.Image), &g.op)
}
