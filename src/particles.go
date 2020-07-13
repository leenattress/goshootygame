package main

import (
	"github.com/hajimehoshi/ebiten"
	"math/rand"
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

func explodeSmall(g *Game, x float64, y float64) {
	// big white flash
	g.particles.particles = append(g.particles.particles, &Particle{
		x:            x,
		y:            y,
		vx:           0,
		vy:           0,
		size:         100,
		sizev:        -10,
		particleType: 2,
		life:         6,
	})
	// smaller fireballs
	for i := 0; i < 8; i++ {
		g.particles.particles = append(g.particles.particles, &Particle{
			x:            x,
			y:            y,
			vx:           float64(4 - rand.Intn(8)),
			vy:           float64(4 - rand.Intn(8)),
			size:         float64(rand.Intn(30) + 20),
			sizev:        -3,
			particleType: 1,
			life:         10,
		})
	}
}

func explodeBig(g *Game, x float64, y float64) {
	// big white flash
	g.particles.particles = append(g.particles.particles, &Particle{
		x:            x,
		y:            y,
		vx:           0,
		vy:           0,
		size:         250,
		sizev:        -10,
		particleType: 2,
		life:         8,
	})
	// smaller fireballs
	for i := 0; i < 20; i++ {
		g.particles.particles = append(g.particles.particles, &Particle{
			x:            x,
			y:            y,
			vx:           float64(4 - rand.Intn(8)),
			vy:           float64(4 - rand.Intn(8)),
			size:         float64(rand.Intn(40) + 30),
			sizev:        -2,
			particleType: 1,
			life:         15,
		})
	}
}
