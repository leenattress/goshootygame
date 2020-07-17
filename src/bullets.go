package main

import ()

// Bullet is our player bullets
type Bullet struct {
	imageWidth  int
	imageHeight int
	x           float64
	y           float64
	vx          float64
	vy          float64
	angle       int
	toDelete    bool
	hitbox      Hitbox
}

//Bullets is an array of bullet
type Bullets struct {
	bullets []*Bullet
	num     int
}

// func bulletExists(arr []*Bullet, index int) bool {
// 	return (len(arr) > index)
// }
