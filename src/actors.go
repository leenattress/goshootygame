package main

import (
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"image/color"
)

// Hitbox is used to determine collisions
type Hitbox struct {
	x float64
	y float64
	w float64
	h float64
}

type Actor struct {
	group       string
	imageWidth  int
	imageHeight int
	x           float64
	y           float64
	vx          float64
	vy          float64
	angle       int
	actorType   string
	sprite      string
	toDelete    bool
	t           int
	hitbox      Hitbox
}

// Actors is an array of Actor and num, to count
type Actors struct {
	actors []*Actor
	num    int
}

// Create an actor
func (a *Actors) Create(newActor Actor) {
	a.actors = append(a.actors, &Actor{
		group:       newActor.group,
		imageWidth:  newActor.imageWidth,
		imageHeight: newActor.imageHeight,
		x:           newActor.x,
		y:           newActor.y,
		vx:          newActor.vx,
		vy:          newActor.vy,
		angle:       newActor.angle,
		actorType:   newActor.actorType,
		sprite:      newActor.sprite,
		toDelete:    false,
		t:           newActor.t,
		hitbox: Hitbox{
			x: newActor.hitbox.x,
			y: newActor.hitbox.y,
			w: newActor.hitbox.w,
			h: newActor.hitbox.h,
		},
	})
	a.num = len(a.actors)
}

// Update runs against every actor
func (a *Actors) Update() {
	for i := 0; i < len(a.actors); i++ {
		a.actors[i].Update()
	}
}

//Clean up actors that are to be deleted
func (a *Actors) Clean() bool {
	var tempActors = make([]*Actor, 0)
	var atLeastOne bool = false
	for _, actor := range a.actors {
		if !actor.toDelete {
			tempActors = append(tempActors, actor)
		} else {
			atLeastOne = true
		}
	}
	a.actors = tempActors
	return atLeastOne
}

// Kill mark the actor to be removed
func (a *Actor) Kill() {
	a.toDelete = true
}

// SetPosition sets position of the actor
func (a *Actor) SetPosition(x float64, y float64) {
	a.x = x
	a.y = y
}

// SetPosition sets position of the actor
func (a *Actor) SetVectors(vx float64, vy float64) {
	a.vx = vx
	a.vy = vy
}

// CollidesHitbox does this hitbox collide with anything in this group?
func (a *Actors) CollidesHitbox(x float64, y float64, hitbox Hitbox, group string) bool {
	var hasCollided bool = false
	for j := len(a.actors) - 1; j >= 0; j-- {
		var b = a.actors[j]

		if b.group == group {
			if collide(
				x+hitbox.x,
				y+hitbox.y,
				hitbox.w,
				hitbox.h,
				b.x+b.hitbox.x,
				b.y+b.hitbox.y,
				b.hitbox.w,
				b.hitbox.h,
			) {
				hasCollided = true
			}
		}
	}
	return hasCollided
}

// Update an Actor
func (a *Actor) Update() {
	var newX = a.x + a.vx
	var newY = a.y + a.vy
	a.SetPosition(newX, newY)

	a.t++ // tick the timer for this actor
}

//Draw this group of actors
func (a *Actors) DrawGroup(g *Game, screen *ebiten.Image, group string) {
	for i := 0; i < len(a.actors); i++ {
		if !a.actors[i].toDelete {
			if group == a.actors[i].group {
				s := a.actors[i]

				var w, h int
				w = g.sprites[s.sprite].width
				h = g.sprites[s.sprite].height

				g.op.GeoM.Reset()
				g.op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
				//g.op.GeoM.Rotate(2 * math.Pi * float64(s.angle) / maxAngle)
				g.op.GeoM.Translate(float64(w)/2, float64(h)/2)
				g.op.GeoM.Translate(float64(s.x), float64(s.y))
				//screen.DrawImage(thisImg, &g.op)
				spriteDraw(screen, g, s.sprite)
				if g.debug {
					ebitenutil.DrawRect(
						screen,
						float64(s.x+s.hitbox.x),
						float64(s.y+s.hitbox.y),
						float64(s.hitbox.w),
						float64(s.hitbox.h),
						color.NRGBA{0xff, 0x00, 0x00, 0x77},
					)
				}
			}
		}
	}
}
