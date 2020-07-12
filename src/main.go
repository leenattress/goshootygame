package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"strconv"
)

// /* xml things */
// var data = []byte(`
// <TextureAtlas imagePath="atlas-1.png">
//     <SubTexture name="circleWhite" x="0" y="0" width="64" height="64"/>
//     <SubTexture name="player" x="64" y="0" width="32" height="32"/>
//     <SubTexture name="enemy1" x="96" y="0" width="32" height="32"/>
//     <SubTexture name="enemy2" x="128" y="0" width="32" height="32"/>
//     <SubTexture name="enemy3" x="160" y="0" width="32" height="32"/>
//     <SubTexture name="lives" x="192" y="0" width="16" height="16"/>
//     <SubTexture name="bullet" x="208" y="0" width="8" height="8"/>
//     <SubTexture name="starFast" x="64" y="32" width="1" height="32"/>
//     <SubTexture name="starSlow" x="65" y="32" width="1" height="32"/>
// </TextureAtlas>
// `)

type Node struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:"-"`
	Content []byte     `xml:",innerxml"`
	Nodes   []Node     `xml:",any"`
}

func (n *Node) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.Attrs = start.Attr
	type node Node

	return d.DecodeElement((*node)(n), &start)
}

func walk(nodes []Node, f func(Node) bool) {
	for _, n := range nodes {
		if f(n) {
			walk(n.Nodes, f)
		}
	}
}

/*

Tasks - Galaga Clone

[x] - detect joystick
[x] - player can move using joystick
[x] - bullets can fire upwards
[x] - bullets are removed when they go out of view
[x] - limit fire rate for player
[x] - enemy ships can be created
[x] - enemy ships can be shot and die
[x] - player has lives
[x] - Migrate to sprite atlas, instead of individual sprites.
[ ] - enemy ships can hit player, and lose a life
[ ] - enemy ships can shoot downwards
[ ] - enemy bullets can hit player, and lose a life
[x] - explosions on all things that need it
[x] - scoring
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
[x] - particles

*/

const (
	screenWidth  = 240
	screenHeight = 320
	maxAngle     = 256
)

var (
	debug       bool = true
	spriteAtlas *ebiten.Image
)

func init() {

}

/** PATHS */
type Vertex struct {
	x int
	y int
}
type Path struct {
	vertices []Vertex
}

type Controls struct {
	up    bool
	down  bool
	left  bool
	right bool
	fire  bool
}
type Sprite struct {
	name   string
	x      int
	y      int
	width  int
	height int
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
	particles      Particles
	difficulty     int
	controls       Controls
	sprites        map[string]Sprite
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
	lives       int
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

/** PARTICLES */
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

type Particles struct {
	particles []*Particle
	num       int
}

func (s *Particles) Update() {
	for i := 0; i < len(s.particles); i++ {
		s.particles[i].Update()
	}
}

func (s *Particle) Update() {

	s.x += s.vx
	s.y += s.vy
	s.speed += s.speedv
	s.size += s.sizev
	if !s.forever {
		s.life -= 1
		if s.life < 0 {
			s.toDelete = true
		}
	}

	// if s.life != -1 {
	// 	s.life -= 1 // life ebbs away for this particle
	// }

	// special behaviour for falling stars, they wrap back to the top once they reach the bottom
	if s.particleType == 0 {
		if s.y > screenHeight {
			s.y = -64
		} // wrap around to top
	}

	s.t += 1 // time ticks on
}

/** /PARTICLES */

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
			log.Println("COLLIDE")
		}
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

	img, _, err := image.Decode(bytes.NewReader(packagepng))
	spriteAtlas, _ = ebiten.NewImageFromImage(img, ebiten.FilterDefault)

	if err != nil {
		log.Fatal(err)
	}

	// prepare the sprite atlas in memory
	buf := bytes.NewBuffer(packagexml)
	dec := xml.NewDecoder(buf)

	var n Node
	errXML := dec.Decode(&n)
	if errXML != nil {
		panic(errXML)
	}

	m := make(map[string]Sprite)

	walk([]Node{n}, func(n Node) bool {
		if n.XMLName.Local == "SubTexture" {
			// fmt.Println(string(n.Content))

			x, err := strconv.Atoi(n.Attrs[1].Value)
			if err != nil {
				panic(err)
			}
			y, err := strconv.Atoi(n.Attrs[2].Value)
			if err != nil {
				panic(err)
			}
			width, err := strconv.Atoi(n.Attrs[3].Value)
			if err != nil {
				panic(err)
			}
			height, err := strconv.Atoi(n.Attrs[4].Value)
			if err != nil {
				panic(err)
			}

			m[n.Attrs[0].Value] = Sprite{
				name:   n.Attrs[0].Value,
				x:      x,
				y:      y,
				width:  width,
				height: height,
			}
		}
		return true
	})

	g.sprites = m

	fmt.Println(g.sprites)

	g.player.x = (screenWidth / 2) - 16
	g.player.y = screenHeight - 50
	g.player.vx = 0
	g.player.vy = 0
	g.player.speed = 2
	g.player.maxSpeed = 4
	g.player.fireRate = 0
	g.player.maxFireRate = 8
	g.player.hitbox.x = 8
	g.player.hitbox.y = 8
	g.player.hitbox.w = 8
	g.player.hitbox.h = 8
	g.player.lives = 3

	// create some star particles
	for i := 0; i < 50; i++ {
		g.particles.particles = append(g.particles.particles, &Particle{
			x:       float64(rand.Intn(screenWidth)),
			y:       float64(rand.Intn(screenHeight)),
			vy:      float64(rand.Intn(10) + 1),
			forever: true,
		})
	}

}

func bulletExists(arr []*Bullet, index int) bool {
	return (len(arr) > index)
}

func collide(x1 float64, y1 float64, w1 float64, h1 float64, x2 float64, y2 float64, w2 float64, h2 float64) bool {
	// Check x and y for overlap
	if x2 > w1+x1 || x1 > w2+x2 || y2 > h1+y1 || y1 > h2+y2 {
		return false
	}
	return true
}

/** GAME MAIN UPDATE */
func (g *Game) Update(screen *ebiten.Image) error {
	if !g.inited {
		g.init()
	}

	g.player.vx = 0
	g.player.vy = 0
	g.controls.right = false
	g.controls.left = false
	g.controls.down = false
	g.controls.up = false
	g.controls.fire = false

	// When the "up arrow key" is pressed..
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.controls.up = true
	}
	// When the "down arrow key" is pressed..
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.controls.down = true
	}
	// When the "left arrow key" is pressed..
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.controls.left = true
	}
	// When the "right arrow key" is pressed..
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.controls.right = true
	}
	// When the "space" is pressed..
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.controls.fire = true
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

		v := ebiten.GamepadAxis(id, 0)
		h := ebiten.GamepadAxis(id, 1)
		if v == 1.0 {
			g.controls.right = true
		}
		if v == -1.0 {
			g.controls.left = true
		}
		if h == 1.0 {
			g.controls.down = true
		}
		if h == -1.0 {
			g.controls.up = true
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
				g.controls.fire = true

				//log.Printf("button pressed: id: %d, button: %d", id, b)
			}
			if inpututil.IsGamepadButtonJustReleased(id, b) {
				//log.Printf("button released: id: %d, button: %d", id, b)
			}
		}

		if g.controls.right {
			g.player.vx = g.player.speed
		}
		if g.controls.left {
			g.player.vx = -g.player.speed
		}
		if g.controls.down {
			g.player.vy = g.player.speed
		}
		if g.controls.up {
			g.player.vy = -g.player.speed
		}
		if g.controls.fire {
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
		}

		//act on movement for player
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
	}
	g.actors.Update(g)

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

				// big shite flash
				g.particles.particles = append(g.particles.particles, &Particle{
					x:            b.x,
					y:            b.y,
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
						x:            b.x,
						y:            b.y,
						vx:           float64(4 - rand.Intn(8)),
						vy:           float64(4 - rand.Intn(8)),
						size:         float64(rand.Intn(30) + 20),
						sizev:        -3,
						particleType: 1,
						life:         10,
					})
				}
			}
		}

		g.bullets.bullets[i].x += g.bullets.bullets[i].vx
		g.bullets.bullets[i].y += g.bullets.bullets[i].vy

		if g.bullets.bullets[i].y < 0 {
			g.bullets.bullets[i].toDelete = true
		}

		if len(g.actors.actors) == 0 {
			// create some baddies
			var thisWave int = rand.Intn(3)
			for i := 0; i < 5; i++ {
				for j := 0; j < 4; j++ {
					g.actors.actors = append(g.actors.actors, &Actor{
						imageWidth:  32,
						imageHeight: 32,
						x:           float64(12 + (i * 40)), // all these ssquares make a circle
						y:           float64(58 + (j * 32)),
						vx:          0,
						vy:          0,
						angle:       0,
						enemyType:   thisWave,
						toDelete:    false,
						t:           (i + j) * 3, // by starting the timer offset like this we get a pleasant wiggly formation
						hitbox: Hitbox{
							x: 4,
							y: 4,
							w: 24,
							h: 24,
						},
					})
				}
			}
			g.difficulty += 1
		}
	}

	var tempBullets = make([]*Bullet, 0)
	for _, x := range g.bullets.bullets {
		if !x.toDelete {
			tempBullets = append(tempBullets, x)
		}
	}
	g.bullets.bullets = tempBullets

	var tempActors = make([]*Actor, 0)
	for _, a := range g.actors.actors {
		if !a.toDelete {
			tempActors = append(tempActors, a)
		}
	}
	g.actors.actors = tempActors

	var tempParticles = make([]*Particle, 0)
	for _, x := range g.particles.particles {
		if !x.toDelete {
			tempParticles = append(tempParticles, x)
		}
	}
	g.particles.particles = tempParticles

	g.particles.Update()

	return nil
}

func spriteDraw(screen *ebiten.Image, g *Game, sprite string) {
	screen.DrawImage(spriteAtlas.SubImage(image.Rect(g.sprites[sprite].x, g.sprites[sprite].y, g.sprites[sprite].x+g.sprites[sprite].width, g.sprites[sprite].y+g.sprites[sprite].height)).(*ebiten.Image), &g.op)
}

/** GAME MAIN DRAW */
func (g *Game) Draw(screen *ebiten.Image) {

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
	//screen.DrawImage(playerImg, &g.op)

	spriteDraw(screen, g, "player")

	w, h := g.sprites["bullet"].width, g.sprites["bullet"].height
	for i := 0; i < len(g.bullets.bullets); i++ {
		if !g.bullets.bullets[i].toDelete {
			s := g.bullets.bullets[i]
			g.op.GeoM.Reset()
			g.op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
			g.op.GeoM.Rotate(2 * math.Pi * float64(s.angle) / maxAngle)
			g.op.GeoM.Translate(float64(w)/2, float64(h)/2)
			g.op.GeoM.Translate(float64(s.x), float64(s.y))
			//screen.DrawImage(bulletImg, &g.op)
			spriteDraw(screen, g, "bullet")
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

			var thisEnemy string
			//var thisImg *ebiten.Image
			var w, h int

			if s.enemyType == 0 {
				thisEnemy = "enemy1"
			}
			if s.enemyType == 1 {
				thisEnemy = "enemy2"
			}
			if s.enemyType == 2 {
				thisEnemy = "enemy3"
			}
			w = g.sprites[thisEnemy].width
			h = g.sprites[thisEnemy].height

			g.op.GeoM.Reset()
			g.op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
			//g.op.GeoM.Rotate(2 * math.Pi * float64(s.angle) / maxAngle)
			g.op.GeoM.Translate(float64(w)/2, float64(h)/2)
			g.op.GeoM.Translate(float64(s.x), float64(s.y))
			//screen.DrawImage(thisImg, &g.op)
			spriteDraw(screen, g, thisEnemy)

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

	for i := 0; i < len(g.particles.particles); i++ {
		s := g.particles.particles[i]

		// stars
		if s.particleType == 0 {
			var scale float64 = s.vy / 9 // magic nine?
			g.op.GeoM.Reset()
			g.op.GeoM.Scale(1, float64(scale))
			g.op.GeoM.Translate(float64(s.x), float64(s.y))
			g.op.ColorM.Translate(0, 0, 0, -(-scale + 1))
			if s.vy > 5 {
				spriteDraw(screen, g, "starFast")
			} else {
				spriteDraw(screen, g, "starSlow")
			}
			g.op.ColorM.Reset()
		}

		// fireballs
		if s.particleType == 1 {
			var scale float64 = s.size / 100 // between 0 and 1
			w, h := g.sprites["circleWhite"].width, g.sprites["circleWhite"].height
			var nudgex float64 = float64(w) * scale
			var nudgey float64 = float64(h) * scale

			g.op.GeoM.Reset()
			g.op.GeoM.Scale(s.size/100, s.size/100)
			g.op.GeoM.Translate(float64(s.x-(nudgex/2)), float64(s.y-(nudgey/2)))

			g.op.ColorM.Translate(2, -scale*2, -1, 0)
			spriteDraw(screen, g, "circleWhite")
			g.op.ColorM.Reset()
		}
		// big white circle
		if s.particleType == 2 {
			var scale float64 = s.size / 100 // between 0 and 1
			w, h := g.sprites["circleWhite"].width, g.sprites["circleWhite"].height
			var nudgex float64 = float64(w) * scale
			var nudgey float64 = float64(h) * scale

			g.op.GeoM.Reset()
			g.op.GeoM.Scale(scale, scale)
			g.op.GeoM.Translate(float64(s.x-(nudgex/2)), float64(s.y-(nudgey/2)))
			spriteDraw(screen, g, "circleWhite")
			g.op.ColorM.Reset()
		}
	}

	for i := 0; i < g.player.lives; i++ {
		g.op.GeoM.Reset()
		g.op.GeoM.Translate(float64(16+(i*18)), float64(screenHeight-20))
		spriteDraw(screen, g, "lives")
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("SCORE: %d  -  WAVE: %d ", g.score*1000, g.difficulty))
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
