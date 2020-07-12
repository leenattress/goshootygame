package main

import (
	"math"
)

// Actor is a structure for enemies
type Actor struct {
	imageWidth  int
	imageHeight int
	x           float64
	y           float64
	vx          float64
	vy          float64
	angle       int
	enemyType   int
	toDelete    bool
	t           int
	hitbox      Hitbox
}

// Actors is an array opf Actor
type Actors struct {
	actors []*Actor
	num    int
}

// Update runs against every actor
func (a *Actors) Update(g *Game) {
	for i := 0; i < len(a.actors); i++ {
		var e = a.actors[i] // enemy
		var p = g.player    // player

		if collide(
			e.x+e.hitbox.x,
			e.y+e.hitbox.y,
			e.hitbox.w,
			e.hitbox.h,
			p.x+p.hitbox.x,
			p.y+p.hitbox.y,
			p.hitbox.w,
			p.hitbox.h,
		) {
			g.player.toDelete = true
		}
		a.actors[i].Update()
	}
}

// Update an Actor
func (a *Actor) Update() {
	//if a.enemyType == 0 {
	a.vx = float64(math.Sin(float64(a.t / 10)))
	a.vy = float64(math.Sin(float64(a.t/20) + 80))
	//}

	a.x += a.vx
	a.y += a.vy

	a.t++ // tick the timer for this actor
}
