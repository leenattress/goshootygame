package main

import (
	"github.com/hajimehoshi/ebiten"
)

// Particle is a simple object that can move long a velocity, grow and shrink, etc. Used in visual effects.
type Particle struct {
	x            float64
	y            float64
	vx           float64
	vy           float64
	size         float64
	sizev        float64
	speed        float64
	speedv       float64
	particleType int
	col1         ebiten.ColorM
	col2         ebiten.ColorM
	col3         ebiten.ColorM
	life         int
	toDelete     bool
	t            int
	forever      bool
}

// Particles are multiple Particle
type Particles struct {
	particles []*Particle
	num       int
}

// Update for every particle
func (s *Particles) Update() {
	for i := 0; i < len(s.particles); i++ {
		s.particles[i].Update()
	}
}

// Update runs once for every particle
func (s *Particle) Update() {

	s.x += s.vx
	s.y += s.vy
	s.speed += s.speedv
	s.size += s.sizev
	if !s.forever {
		s.life--
		if s.life < 0 {
			s.toDelete = true
		}
	}

	// special behaviour for falling stars, they wrap back to the top once they reach the bottom
	if s.particleType == 0 {
		if s.y > screenHeight {
			s.y = -64
		} // wrap around to top
	}

	s.t++ // time ticks on
}
