package main

import (
	"github.com/hajimehoshi/ebiten"
	"image"
	"math"
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

// ldX Length Direction x is used to calculate the x given the length and direction
func ldX(len float64, dir float64) float64 {
	return math.Cos(dir) * len
}

// ldY Length Direction y is used to calculate the y given the length and direction
func ldY(len float64, dir float64) float64 {
	return math.Sin(dir) * len
}
func pointAng(cx float64, cy float64, ex float64, ey float64) float64 {
	return math.Atan2(cx-ex, cy-ey)
}
func pointDist(x1 float64, y1 float64, x2 float64, y2 float64) float64 {
	var x3 float64 = math.Abs(x2 - x1)
	var y3 float64 = math.Abs(y2 - y1)
	return math.Sqrt((x3 * x3) + (y3 * y3))
}
