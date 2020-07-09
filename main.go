package main

import (
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"strconv"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
)

/*

Tasks - Galaga Clone

[x] - detect joystick
[x] - player can move using joystick
[x] - bullets can fire upwards
[x] - bullets are removed when they go out of view
[x] - limit fire rate for player
[x] - enemy ships can be created
[x] - enemy ships can be shot and die
[ ] - player has lives
[ ] - enemy ships can hit player, and lose a life
[ ] - enemy ships can shoot downwards
[ ] - enemy bullets can hit player, and lose a life
[ ] - explosions on all things that need it
[ ] - scoring
[ ] - fabulous ui
[ ] - title screen
[ ] - game over screen
[ ] - high scores storage
[ ] - high scores name joystick entry
[ ] - high scores screen
[ ] - high score
[ ] - sound effects for movement
[ ] - sound effects for shoot
[ ] - sound effects for enemy die
[ ] - sound effects for player die
[ ] - screen shake
[ ] - particles

*/

const (
	screenWidth  = 240
	screenHeight = 320
	maxAngle     = 256
)

var (
	debug       bool = false
	bulletImg   *ebiten.Image
	playerImg   *ebiten.Image
	starSlowImg *ebiten.Image
	starFastImg *ebiten.Image
	enemy1      *ebiten.Image
	enemy2      *ebiten.Image
	enemy3      *ebiten.Image
	circleWhite *ebiten.Image
)

func init() {

	// player image
	var err error
	playerImg, _, err = ebitenutil.NewImageFromFile("assets/player.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	bulletImg, _, err = ebitenutil.NewImageFromFile("assets/bullet.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	starSlowImg, _, err = ebitenutil.NewImageFromFile("assets/starSlow.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}
	starFastImg, _, err = ebitenutil.NewImageFromFile("assets/starFast.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	enemy1, _, err = ebitenutil.NewImageFromFile("assets/enemy1.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	enemy2, _, err = ebitenutil.NewImageFromFile("assets/enemy2.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	enemy3, _, err = ebitenutil.NewImageFromFile("assets/enemy3.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	circleWhite, _, err = ebitenutil.NewImageFromFile("assets/circleWhite.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}
}

/** PATHS */
type Vertex struct {
	x int
	y int
}
type Path struct {
	vertices []Vertex
}

type Game struct {
	gamepadIDs     map[int]struct{}
	axes           map[int][]string
	pressedButtons map[int][]string
	actors         Actors
	op             ebiten.DrawImageOptions
	inited         bool
	player         Player
	bullets        Bullets
	paths          []Path
	score          int
	stars          Stars
}

type Hitbox struct {
	x float64
	y float64
	w float64
	h float64
}
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
}

/** BULLETS */
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
type Bullets struct {
	bullets []*Bullet
	num     int
}

/** /BULLETS */

/** STARS */
type Star struct {
	x     float64
	y     float64
	speed float64
}

type Stars struct {
	stars []*Star
	num   int
}

func (s *Stars) Update() {
	for i := 0; i < len(s.stars); i++ {
		s.stars[i].Update()
	}
}

func (s *Star) Update() {
	s.y += s.speed // always downwards

	if s.y > screenHeight {
		s.y = -64
	} // wrap around to top
}

/** /STARS */

/** Actors (enemies) */
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
type Actors struct {
	actors []*Actor
	num    int
}

func (a *Actors) Update() {
	for i := 0; i < len(a.actors); i++ {
		a.actors[i].Update()
	}
}
func (a *Actor) Update() {
	//if a.enemyType == 0 {
	a.vx = float64(math.Sin(float64(a.t / 10)))
	a.vy = float64(math.Sin(float64(a.t/20) + 80))
	//}

	a.x += a.vx
	a.y += a.vy

	a.t += 1 // tick the timer for this actor
}

/** /Actors */

func (g *Game) init() {
	defer func() {
		g.inited = true
	}()

	// define some vertex paths for enemies to follow
	g.paths = make([]Path, 0)
	g.paths = append(g.paths, Path{
		[]Vertex{
			{
				x: 50,
				y: 100,
			},
			{
				x: 100,
				y: 220,
			},
			{
				x: 95,
				y: 150,
			},
		}})

	g.player.x = (screenWidth / 2) - 16
	g.player.y = screenHeight - 50
	g.player.vx = 0
	g.player.vy = 0
	g.player.speed = 2
	g.player.maxSpeed = 4
	g.player.fireRate = 0
	g.player.maxFireRate = 4
	g.player.hitbox.x = 8
	g.player.hitbox.y = 8
	g.player.hitbox.w = 8
	g.player.hitbox.h = 8

	// create some stars
	for i := 0; i < 50; i++ {
		g.stars.stars = append(g.stars.stars, &Star{
			x:     float64(rand.Intn(screenWidth)),
			y:     float64(rand.Intn(screenHeight)),
			speed: float64(rand.Intn(10) + 1),
		})
	}

	// create some baddies
	for i := 0; i < 5; i++ {
		for j := 0; j < 3; j++ {
			g.actors.actors = append(g.actors.actors, &Actor{
				imageWidth:  32,
				imageHeight: 32,
				x:           float64(24 + (i * 36)),
				y:           float64(58 + (j * 36)),
				vx:          0,
				vy:          0,
				angle:       0,
				enemyType:   1,
				toDelete:    false,
				t:           (i + j) * 3,
				hitbox: Hitbox{
					x: 4,
					y: 4,
					w: 24,
					h: 24,
				},
			})
		}
	}

}

func bulletExists(arr []*Bullet, index int) bool {
	return (len(arr) > index)
}

// func hitBox( source, target ) {
// 	/* Source and target objects contain x, y and width, height */
// 	return !(
// 		( ( source.y + source.height ) < ( target.y ) ) ||
// 		( source.y > ( target.y + target.height ) ) ||
// 		( ( source.x + source.width ) < target.x ) ||
// 		( source.x > ( target.x + target.width ) )
// 	);
// }
func collide(x1 float64, y1 float64, w1 float64, h1 float64, x2 float64, y2 float64, w2 float64, h2 float64) bool {
	// Check x and y for overlap
	if x2 > w1+x1 || x1 > w2+x2 || y2 > h1+y1 || y1 > h2+y2 {
		return false
	}
	return true
}

// func collide(ax1 float64, ay1 float64, ax2 float64, ay2 float64, bx1 float64, by1 float64, bx2 float64, by2 float64) bool {
// 	var c1 = (ay1 >= by2)
// 	var c2 = (ay2 <= by1)
// 	var c3 = (ax1 <= bx2)
// 	var c4 = (ax2 >= bx1)

// 	if c1 && c2 && c3 && c4 {
// 		log.Printf("%t %t %t %t", c1, c2, c3, c4)

// 		return true
// 	} else {
// 		return false
// 	}
// }

/** GAME MAIN UPDATE */
func (g *Game) Update(screen *ebiten.Image) error {
	if !g.inited {
		g.init()
	}
	if g.gamepadIDs == nil {
		g.gamepadIDs = map[int]struct{}{}
	}

	// Log the gamepad connection eventa.
	for _, id := range inpututil.JustConnectedGamepadIDs() {
		log.Printf("gamepad connected: id: %d", id)
		g.gamepadIDs[id] = struct{}{}
	}
	for id := range g.gamepadIDs {
		if inpututil.IsGamepadJustDisconnected(id) {
			log.Printf("gamepad disconnected: id: %d", id)
			delete(g.gamepadIDs, id)
		}
	}

	g.axes = map[int][]string{}
	g.pressedButtons = map[int][]string{}
	for id := range g.gamepadIDs {

		maxAxis := ebiten.GamepadAxisNum(id)

		g.player.vx = 0
		g.player.vy = 0

		v := ebiten.GamepadAxis(id, 0)
		h := ebiten.GamepadAxis(id, 1)
		if v == 1.0 {
			g.player.vx = g.player.speed
		}
		if v == -1.0 {
			g.player.vx = -g.player.speed
		}
		if h == 1.0 {
			g.player.vy = g.player.speed
		}
		if h == -1.0 {
			g.player.vy = -g.player.speed
		}

		//act on velocity
		g.player.x += g.player.vx
		g.player.y += g.player.vy

		// screen edges for player
		if g.player.x > screenWidth-32 {
			g.player.x = screenWidth - 32
		}
		if g.player.x < 0 {
			g.player.x = 0
		}
		if g.player.y > screenHeight-32 {
			g.player.y = screenHeight - 32
		}
		if g.player.y < 0 {
			g.player.y = 0
		}

		// limit fire rate
		if g.player.fireRate > 0 {
			g.player.fireRate -= 1
		}

		for a := 0; a < maxAxis; a++ {
			v := ebiten.GamepadAxis(id, a)
			g.axes[id] = append(g.axes[id], fmt.Sprintf("%d:%0.2f", a, v))
		}
		maxButton := ebiten.GamepadButton(ebiten.GamepadButtonNum(id))
		for b := ebiten.GamepadButton(id); b < maxButton; b++ {
			if ebiten.IsGamepadButtonPressed(id, b) {
				g.pressedButtons[id] = append(g.pressedButtons[id], strconv.Itoa(int(b)))
			}

			// Log button eventa.
			if inpututil.IsGamepadButtonJustPressed(id, b) {

				if g.player.fireRate == 0 {
					g.player.fireRate = g.player.maxFireRate
					g.bullets.bullets = append(g.bullets.bullets, &Bullet{
						imageWidth:  8,
						imageHeight: 8,
						x:           g.player.x + 12, // bullet spawn at nose
						y:           g.player.y + 4,
						vx:          0,
						vy:          -6,
						angle:       0,
						hitbox: Hitbox{
							x: 0,
							y: 0,
							w: 8,
							h: 8,
						},
					})
					g.bullets.num = len(g.bullets.bullets)
				}

				//log.Printf("button pressed: id: %d, button: %d", id, b)
			}
			if inpututil.IsGamepadButtonJustReleased(id, b) {
				//log.Printf("button released: id: %d, button: %d", id, b)
			}
		}
	}
	g.actors.Update()

	for i := len(g.bullets.bullets) - 1; i >= 0; i-- {
		var b = g.bullets.bullets[i]
		for j := len(g.actors.actors) - 1; j >= 0; j-- {
			var a = g.actors.actors[j]

			if collide(
				a.x+a.hitbox.x,
				a.y+a.hitbox.y,
				a.hitbox.w,
				a.hitbox.h,
				b.x+b.hitbox.x,
				b.y+b.hitbox.y,
				b.hitbox.w,
				b.hitbox.h,
			) {
				g.bullets.bullets[i].toDelete = true
				g.actors.actors[j].toDelete = true
				g.score += 1
			}
		}

		//if bulletExists(g.bullets.bullets, i) {
		//for i := 0; i < g.bullets.num; i++ {
		g.bullets.bullets[i].x += g.bullets.bullets[i].vx
		g.bullets.bullets[i].y += g.bullets.bullets[i].vy

		if g.bullets.bullets[i].y < 0 {
			g.bullets.bullets[i].toDelete = true
		}
		//}
	}

	var tempBullets = make([]*Bullet, 0)
	for _, x := range g.bullets.bullets {
		if !x.toDelete {
			tempBullets = append(tempBullets, x)
		}
	}
	g.bullets.bullets = tempBullets

	var tempActors = make([]*Actor, 0)
	for _, x := range g.actors.actors {
		if !x.toDelete {
			tempActors = append(tempActors, x)
		}
	}
	g.actors.actors = tempActors

	g.stars.Update()

	return nil
}

/** GAME MAIN DRAW */
func (g *Game) Draw(screen *ebiten.Image) {

	// stars are in the background
	for i := 0; i < len(g.stars.stars); i++ {
		s := g.stars.stars[i]
		var scale float64 = s.speed / 9
		g.op.GeoM.Reset()
		g.op.GeoM.Scale(1, float64(scale))
		// g.op.GeoM.Rotate(2 * math.Pi * float64(s.angle) / maxAngle)
		// g.op.GeoM.Translate(float64(w)/2, float64(h)/2)
		g.op.GeoM.Translate(float64(s.x), float64(s.y))
		g.op.ColorM.Translate(0, 0, 0, -(-scale + 1))
		if s.speed > 5 {
			screen.DrawImage(starFastImg, &g.op)
		} else {
			screen.DrawImage(starSlowImg, &g.op)
		}
		g.op.ColorM.Reset()

	}

	// Draw the current gamepad statua.
	str := ""
	if debug {
		if len(g.gamepadIDs) > 0 {

			// for id := range g.gamepadIDs {
			// 	str += fmt.Sprintf("Gamepad (ID: %d, SDL ID: %s):\n", id, ebiten.GamepadSDLID(id))
			// 	str += fmt.Sprintf("  Axes:    %s\n", strings.Join(g.axes[id], ", "))
			// 	str += fmt.Sprintf("  Buttons: %s\n", strings.Join(g.pressedButtons[id], ", "))
			// 	str += fmt.Sprintf("Bullets: %d\n", len(g.bullets.bullets))
			// 	str += fmt.Sprintf("Player: x-%d y-%d\n", int(g.player.x), int(g.player.y))
			// 	str += "\n"
			// }

		} else {
			str = "Please connect your gamepad."
		}
		ebitenutil.DebugPrint(screen, str)
	}
	// draw player sprite
	g.op.GeoM.Reset()
	//g.op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
	//g.op.GeoM.Rotate(2 * math.Pi * float64(s.angle) / maxAngle)
	//g.op.GeoM.Translate(float64(w)/2, float64(h)/2)
	g.op.GeoM.Translate(float64(g.player.x), float64(g.player.y))
	screen.DrawImage(playerImg, &g.op)

	w, h := bulletImg.Size()
	for i := 0; i < len(g.bullets.bullets); i++ {
		if !g.bullets.bullets[i].toDelete {
			s := g.bullets.bullets[i]
			g.op.GeoM.Reset()
			g.op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
			g.op.GeoM.Rotate(2 * math.Pi * float64(s.angle) / maxAngle)
			g.op.GeoM.Translate(float64(w)/2, float64(h)/2)
			g.op.GeoM.Translate(float64(s.x), float64(s.y))
			screen.DrawImage(bulletImg, &g.op)
			if debug {
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

	for i := 0; i < len(g.actors.actors); i++ {
		if !g.actors.actors[i].toDelete {
			s := g.actors.actors[i]

			var thisImg *ebiten.Image
			var w, h int

			if s.enemyType == 0 {
				w = 32
				h = 32
				thisImg = enemy1
			}
			if s.enemyType == 1 {
				w = 32
				h = 32
				thisImg = enemy2
			}
			if s.enemyType == 2 {
				w = 32
				h = 32
				thisImg = enemy3
			}

			g.op.GeoM.Reset()
			g.op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
			g.op.GeoM.Rotate(2 * math.Pi * float64(s.angle) / maxAngle)
			g.op.GeoM.Translate(float64(w)/2, float64(h)/2)
			g.op.GeoM.Translate(float64(s.x), float64(s.y))
			screen.DrawImage(thisImg, &g.op)
			if debug {
				ebitenutil.DrawRect(
					screen,
					float64(s.x+s.hitbox.x),
					float64(s.y+s.hitbox.y),
					float64(s.hitbox.w),
					float64(s.hitbox.h),
					color.NRGBA{0xff, 0x00, 0x00, 0x77},
				)
			}

			ebitenutil.DebugPrint(screen, str)
		}
	}

	// draw path debug
	// var thisPath = g.paths[0]
	// if len(thisPath.vertices) > 1 {
	// 	for i := 0; i < len(thisPath.vertices); i++ {
	// 		if i > 0 {
	// 			ebitenutil.DrawLine(
	// 				screen,
	// 				float64(thisPath.vertices[i].x),
	// 				float64(thisPath.vertices[i].y),
	// 				float64(thisPath.vertices[i-1].x),
	// 				float64(thisPath.vertices[i-1].y),
	// 				color.NRGBA{0x00, 0x00, 0xff, 0xff},
	// 			)
	// 		}
	// 	}
	// }

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {

	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Game Window")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
