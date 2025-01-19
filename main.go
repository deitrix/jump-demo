package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func main() {
	game := &Game{
		Player: Player{
			Pos: Vec2{X: mapSize * tileSize / 2, Y: mapSize * tileSize / 2},
		},
		Map: make([][]int, 0),
	}
	// Initialize the map
	for i := 0; i < mapSize; i++ {
		game.Map = append(game.Map, make([]int, mapSize))
	}
	// Add the floor
	for i := 0; i < mapSize; i++ {
		game.Map[mapSize-1][i] = 1
	}

	ebiten.SetWindowSize(640, 640)
	ebiten.SetWindowTitle("Jump Demo")
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}

const (
	mapSize  = 16
	tileSize = 16
)

type Game struct {
	Player Player
	Map    [][]int
}

const playerSpeed = 4

func (g *Game) Update() error {
	x, y := ebiten.CursorPosition()
	tileX, tileY := x/tileSize, y/tileSize
	if tileX >= 0 && tileX < mapSize && tileY >= 0 && tileY < mapSize {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.Map[tileY][tileX] = 1
		}
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			g.Map[tileY][tileX] = 0
		}
	}

	// If AD is pressed, move the player by inducing velocity
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.Player.Vel.X = -playerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		g.Player.Vel.X = playerSpeed
	}
	// If Space is pressed and the player is touching the ground, jump
	if !g.Player.Falling() && inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.Player.Vel.Y = -playerSpeed * 2.5
	}

	// Apply gravity
	g.Player.Vel.Y += .5

	// Decay the player's velocity. This stops the player's acceleration from growing indefinitely.
	// This is a simple way to simulate friction and air resistance.
	g.Player.Vel.X *= 0.7
	g.Player.Vel.Y *= 0.97

	// Update the player's position
	g.Player.LastPos = g.Player.Pos
	g.Player.Pos.MAdd(g.Player.Vel)

	// Check for collisions with the map
	playerRect := Rect{X: g.Player.Pos.X - playerWidth/2, Y: g.Player.Pos.Y - playerHeight, Width: playerWidth, Height: playerHeight}

	// Check for collisions with the map
	for y, row := range g.Map {
		for x, tile := range row {
			if tile == 0 {
				continue
			}
			tileRect := Rect{X: float64(x * tileSize), Y: float64(y * tileSize), Width: tileSize, Height: tileSize}
			collision, resolve := checkCollision(playerRect, tileRect)
			if collision {
				switch {
				case abs(resolve.X) < abs(resolve.Y):
					g.Player.Pos.X += resolve.X
					// Only set the velocity to 0 if the player is moving into the tile.
					if g.Player.Vel.X > 0 && resolve.X < 0 || g.Player.Vel.X < 0 && resolve.X > 0 {
						g.Player.Vel.X = 0
					}
				case abs(resolve.X) > abs(resolve.Y):
					g.Player.Pos.Y += resolve.Y
					// Only set the velocity to 0 if the player is moving into the tile.
					if g.Player.Vel.Y > 0 && resolve.Y < 0 || g.Player.Vel.Y < 0 && resolve.Y > 0 {
						g.Player.Vel.Y = 0
					}
				}
				// Update the player rect, as it will have changed.
				playerRect = Rect{X: g.Player.Pos.X - playerWidth/2, Y: g.Player.Pos.Y - playerHeight, Width: playerWidth, Height: playerHeight}
			}
		}
	}

	return nil
}

// checkCollision checks if two rectangles are colliding and returns a boolean indicating if they are colliding
// and a vector to resolve the collision. The resolve vector is the minimum translation vector needed to separate
// the two rectangles.
func checkCollision(a, b Rect) (collision bool, resolve Vec2) {
	axlo, axhi, bxlo, bxhi := a.X, a.X+a.Width, b.X, b.X+b.Width
	aylo, ayhi, bylo, byhi := a.Y, a.Y+a.Height, b.Y, b.Y+b.Height

	xOverlap := axlo < bxhi && axhi > bxlo
	yOverlap := aylo < byhi && ayhi > bylo

	if xOverlap && yOverlap {
		collision = true
		if axlo < bxlo {
			resolve.X = bxlo - axhi
		} else {
			resolve.X = bxhi - axlo
		}
		if aylo < bylo {
			resolve.Y = bylo - ayhi
		} else {
			resolve.Y = byhi - aylo
		}
	}

	return collision, resolve
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Draw the map
	for y, row := range g.Map {
		for x, tile := range row {
			if tile == 0 {
				continue
			}
			vector.DrawFilledRect(screen, float32(x*tileSize), float32(y*tileSize), tileSize, tileSize, color.White, false)
		}
	}

	// Draw the cursor
	x, y := ebiten.CursorPosition()
	tileX, tileY := x/tileSize, y/tileSize
	if tileX >= 0 && tileX < mapSize && tileY >= 0 && tileY < mapSize {
		col := color.RGBA{255, 255, 255, 255}
		if g.Map[tileY][tileX] == 1 {
			col = color.RGBA{255, 0, 0, 255}
		}
		vector.StrokeRect(screen, float32(tileX*tileSize)+1, float32(tileY*tileSize)+1, tileSize-1, tileSize-1, 1, col, false)
	}

	// Draw the player
	vector.DrawFilledRect(
		screen,
		float32(g.Player.Pos.X)-playerWidth/2,
		float32(g.Player.Pos.Y)-playerHeight,
		playerWidth,
		playerHeight,
		color.RGBA{0, 117, 234, 255},
		false,
	)
}

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return mapSize * tileSize, mapSize * tileSize
}

const (
	playerWidth  = 16
	playerHeight = 32
)

type Rect struct {
	X, Y, Width, Height float64
}

type Player struct {
	Pos     Vec2
	LastPos Vec2
	Vel     Vec2
}

func (p Player) Falling() bool {
	return p.Pos.Y < p.LastPos.Y
}

type Vec2 struct {
	X, Y float64
}

func (v *Vec2) MAdd(other Vec2) {
	v.X += other.X
	v.Y += other.Y
}

func (v *Vec2) MSub(other Vec2) {
	v.X -= other.X
	v.Y -= other.Y
}

func (v *Vec2) MMul(scalar float64) {
	v.X *= scalar
	v.Y *= scalar
}

func (v *Vec2) MDiv(scalar float64) {
	v.X /= scalar
	v.Y /= scalar
}

func (v *Vec2) Add(other Vec2) Vec2 {
	return Vec2{X: v.X + other.X, Y: v.Y + other.Y}
}

func (v *Vec2) Sub(other Vec2) Vec2 {
	return Vec2{X: v.X - other.X, Y: v.Y - other.Y}
}

func (v *Vec2) Mul(scalar float64) Vec2 {
	return Vec2{X: v.X * scalar, Y: v.Y * scalar}
}

func (v *Vec2) Div(scalar float64) Vec2 {
	return Vec2{X: v.X / scalar, Y: v.Y / scalar}
}

func abs(a float64) float64 {
	if a < 0 {
		return -a
	}
	return a
}
