package main

import ()

// Player is the player state object
type Player struct {
	x           float64
	y           float64
	vx          float64
	vy          float64
	speed       float64
	maxSpeed    float64
	fireRate    int16
	maxFireRate int16
	hitbox      Hitbox
	lives       int
	toDelete    bool
	safety      int
}

func newPlayer() Player {
	return Player{
		x:           (screenWidth / 2) - 16,
		y:           screenHeight - 50,
		vx:          0,
		vy:          0,
		speed:       2,
		maxSpeed:    4,
		fireRate:    0,
		maxFireRate: 8,
		hitbox: Hitbox{
			x: 8,
			y: 8,
			w: 8,
			h: 8,
		},
		safety: 120,
	}
}
