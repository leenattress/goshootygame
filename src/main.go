package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/audio"
	"github.com/hajimehoshi/ebiten/audio/mp3"
	"github.com/hajimehoshi/ebiten/audio/wav"
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
[x] - Pack all the assets into the executable.
[x] - Temporary immunity on spawn
[x] - enemy ships can hit player, and lose a life
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
	sampleRate   = 44100
)

var (
	debug              bool = false
	spriteAtlas        *ebiten.Image
	t                  int
	sfxVolume          float64 = 0.4
	bgmVolume          float64 = 0.3
	audioContext       *audio.Context
	audioShooty        *audio.Player
	audioDeath         *audio.Player
	audioPlayerExplode *audio.Player
	audioExploded      *audio.Player
	audioMusic         *audio.Player
)

func init() {
	var err error
	audioContext, err = audio.NewContext(sampleRate)
	if err != nil {
		log.Fatal(err)
	}
}

// Controls describes the current state of the input
type Controls struct {
	up    bool
	down  bool
	left  bool
	right bool
	fire  bool
}

// Sprite is used in the construction of the sprite atlas object
type Sprite struct {
	name   string
	x      int
	y      int
	width  int
	height int
}

// Game is the state of our game
type Game struct {
	time           int
	gamepadIDs     map[int]struct{}
	axes           map[int][]string
	pressedButtons map[int][]string
	actors         Actors
	op             ebiten.DrawImageOptions
	inited         bool
	player         Player
	bullets        Bullets
	score          int
	particles      Particles
	difficulty     int
	controls       Controls
	sprites        map[string]Sprite
	enemyShoot     int
	lives          int
	debug          bool
}

func (g *Game) init() {
	g.debug = debug
	defer func() {
		g.inited = true
	}()

	// get the sprite atlas from code
	img, _, err := image.Decode(bytes.NewReader(packagepng))
	spriteAtlas, _ = ebiten.NewImageFromImage(img, ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	// get a shooty sound
	audioShootDecoded, err := wav.Decode(audioContext, audio.BytesReadSeekCloser(shootSample))
	audioShooty, err = audio.NewPlayer(audioContext, audioShootDecoded)
	audioShooty.SetVolume(sfxVolume - 0.2)
	if err != nil {
		log.Fatal(err)
	}

	// get a player death sound
	audioDeathDecoded, err := wav.Decode(audioContext, audio.BytesReadSeekCloser(deathSample))
	audioDeath, err = audio.NewPlayer(audioContext, audioDeathDecoded)
	audioDeath.SetVolume(sfxVolume + 0.3)
	if err != nil {
		log.Fatal(err)
	}

	// get a small explosion sound
	audioExplodeDecoded, err := wav.Decode(audioContext, audio.BytesReadSeekCloser(explodeSample))
	audioExploded, err = audio.NewPlayer(audioContext, audioExplodeDecoded)
	audioExploded.SetVolume(sfxVolume - 0.2)
	if err != nil {
		log.Fatal(err)
	}

	// get background music
	audioMusicDecoded, err := mp3.Decode(audioContext, audio.BytesReadSeekCloser(bgmSample))
	audioMusic, err = audio.NewPlayer(audioContext, audioMusicDecoded)
	audioMusic.SetVolume(bgmVolume)
	if err != nil {
		log.Fatal(err)
	}
	audioMusic.Rewind()
	audioMusic.Play()

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

	g.enemyShoot = 120 // start our enemies shooting

	g.player = newPlayer()
	g.lives = 3
	g.player.safety = 60 * 4

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

// Update the Game object
func (g *Game) Update(screen *ebiten.Image) error {
	if !g.inited {
		g.init()
	}
	t++

	if g.enemyShoot == 0 && len(g.actors.actors) > 0 {

		var enemyToShoot = rand.Intn(len(g.actors.actors))
		var actorSprite string = "enemyBullet"
		g.actors.Create(Actor{
			group:       "enemyBullet",
			imageWidth:  g.sprites[actorSprite].width,
			imageHeight: g.sprites[actorSprite].height,
			x:           g.actors.actors[enemyToShoot].x, // all these squares make a circle
			y:           g.actors.actors[enemyToShoot].y,
			vx:          0,
			vy:          3,
			actorType:   "bullet",
			sprite:      actorSprite,
			hitbox: Hitbox{
				x: 0,
				y: 0,
				w: float64(g.sprites[actorSprite].width),
				h: float64(g.sprites[actorSprite].height),
			},
		})

		g.enemyShoot = 60 //rand.Intn(30) + 30 - (g.difficulty * 2)
		if g.enemyShoot < 10 {
			g.enemyShoot = 10
		}

	} else {
		if g.enemyShoot > 0 {
			g.enemyShoot--
		}
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
			audioShooty.Rewind()
			audioShooty.Play()
		}
	}

	// only move and shoot if alive
	if g.player.alive {

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
			g.player.fireRate--
		}

		// engine trail
		g.particles.particles = append(g.particles.particles, &Particle{
			x:            g.player.x,
			y:            g.player.y,
			vx:           0,
			vy:           0,
			size:         10,
			sizev:        -1,
			particleType: 99,
			life:         6,
		})

	}

	// Update the vectors
	for i := len(g.actors.actors) - 1; i >= 0; i-- {
		var a = g.actors.actors[i]
		if a.group == "enemy" {
			var vx = float64(math.Sin(float64(a.t / 10)))
			var vy = float64(math.Sin(float64(a.t/20) + 80))
			if a.actorType == "enemy1" || a.actorType == "enemy2" || a.actorType == "enemy3" {
				a.SetVectors(vx, vy)
			}

		}
		if a.group == "enemyBullet" {
			if a.y > screenHeight {
				a.Kill()
			}
		}
	}
	g.actors.Update()

	for i := len(g.bullets.bullets) - 1; i >= 0; i-- {
		var b = g.bullets.bullets[i]
		for j := len(g.actors.actors) - 1; j >= 0; j-- {
			var a = g.actors.actors[j]

			if a.actorType != "5" {
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
					g.actors.actors[j].Kill()
					g.score++
					explodeSmall(g, b.x, b.y)
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
					g.actors.Create(Actor{
						group:       "enemy",
						actorType:   "enemy" + strconv.Itoa(thisWave),
						sprite:      "enemy" + strconv.Itoa(thisWave),
						imageWidth:  32,
						imageHeight: 32,
						x:           float64(12 + (i * 40)), // all these squares make a circle
						y:           float64(48 + (j * 32)),
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
			g.difficulty++
		}
	}

	var tempBullets = make([]*Bullet, 0)
	for _, x := range g.bullets.bullets {
		if !x.toDelete {
			tempBullets = append(tempBullets, x)
		}
	}
	g.bullets.bullets = tempBullets

	var deletedAny = g.actors.Clean()
	if deletedAny {
		audioExploded.Rewind()
		audioExploded.Play()
	}

	var tempParticles = make([]*Particle, 0)
	for _, x := range g.particles.particles {
		if !x.toDelete {
			tempParticles = append(tempParticles, x)
		}
	}
	g.particles.particles = tempParticles
	g.particles.Update()

	// Does the player collide with any enemy?
	if g.actors.CollidesHitbox(g.player.x, g.player.y, g.player.hitbox, "enemy") && g.player.safety <= 0 {
		g.player.toDelete = true
	}
	// Does the player collide with any enemy bullets??
	if g.actors.CollidesHitbox(g.player.x, g.player.y, g.player.hitbox, "enemyBullet") && g.player.safety <= 0 {
		g.player.toDelete = true
	}

	if g.player.toDelete {
		killPlayer(g)
	}

	return nil
}

// Draw is called every frame to draw the game contents
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

	if g.player.alive {
		// draw player sprite
		g.op.GeoM.Reset()
		//g.op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
		//g.op.GeoM.Rotate(2 * math.Pi * float64(s.angle) / maxAngle)
		//g.op.GeoM.Translate(float64(w)/2, float64(h)/2)
		g.op.GeoM.Translate(float64(g.player.x), float64(g.player.y))
		//screen.DrawImage(playerImg, &g.op)

		if g.player.safety > 0 {
			if t%2 == 0 {
				// flicker is safe
				spriteDraw(screen, g, "player")
			}
		} else {
			spriteDraw(screen, g, "player")
		}
		// some rotating stars
		wp, hp := g.sprites["player"].width, g.sprites["player"].height       // player width/height
		ws, hs := g.sprites["starSmall"].width, g.sprites["starSmall"].height // small star
		wst, hst := g.sprites["starTiny"].width, g.sprites["starTiny"].height // tiny star

		if g.player.safety > 0 {

			for i := 0; i < 6; i++ {
				g.op.GeoM.Reset()
				g.op.GeoM.Translate(-float64(ws)/2, -float64(hs)/2) // center this sprite
				g.op.GeoM.Translate(float64(wp)/2, float64(hp)/2)   // center on player
				g.op.GeoM.Translate(
					float64(
						g.player.x+ldX(
							24, float64(t+(i*10))/10,
						),
					),
					float64(
						g.player.y+ldY(
							24, float64(t+(i*10))/10,
						),
					),
				)
				spriteDraw(screen, g, "starSmall")

				g.op.GeoM.Reset()
				g.op.GeoM.Translate(-float64(wst)/2, -float64(hst)/2)
				g.op.GeoM.Translate(float64(wp)/2, float64(hp)/2)
				g.op.GeoM.Translate(
					float64(
						g.player.x+ldX(
							32, -float64(t+(i*18))/18,
						),
					),
					float64(
						g.player.y+ldY(
							32, -float64(t+(i*18))/18,
						),
					),
				)
				spriteDraw(screen, g, "starTiny")
			}
			g.player.safety--
		}

	}

	g.actors.DrawGroup(g, screen, "enemy")
	g.actors.DrawGroup(g, screen, "enemyBullet")

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

	for i := 0; i < len(g.particles.particles); i++ {
		s := g.particles.particles[i]

		// stars
		if s.particleType == 0 {
			var scale float64 = s.vy / 9 // magic nine?
			g.op.GeoM.Reset()
			g.op.GeoM.Scale(1, float64(scale))
			g.op.GeoM.Translate(float64(s.x), float64(s.y))
			g.op.ColorM.Translate(0, 0, 0, -(0.5 + (-scale + 1)))
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

	for i := 0; i < g.lives; i++ {
		g.op.GeoM.Reset()
		g.op.GeoM.Translate(float64(16+(i*18)), float64(screenHeight-20))
		spriteDraw(screen, g, "lives")
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("SCORE: %d  -  WAVE: %d ", g.score*1000, g.difficulty))
}

// Layout is part of the ebiten framework
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
