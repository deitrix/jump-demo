package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func main() {
	game := &Game{
		CurrentPlayer: Yellow,
	}

	ebiten.SetWindowSize(1280, 960)
	ebiten.SetWindowTitle("Connect4")
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}

type Game struct {
	Slots         [7][6]Piece
	CurrentPlayer Piece
	Falling       bool    // Is a piece currently falling?
	VelocityY     float64 // The velocity of the falling piece
	FallingX      float64 // The x position of the falling piece, in pixels
	FallingY      float64 // The y position of the falling piece, in pixels
	FallStopY     float64 // The y position, in pixels, where the falling piece should stop
	TargetX       int     // The target x position of the falling piece
	TargetY       int     // The target y position of the falling piece
}

type Piece uint

const (
	_ Piece = iota
	Yellow
	Red
)

func (g *Game) Update() error {
	mx, _ := ebiten.CursorPosition()
	slotX := min(max((mx-boardPaddingPx)/(slotSizePx+slotPaddingPx), 0), 6)
	if !g.Falling && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		for y := 0; y < 6; y++ {
			if g.Slots[slotX][y] == 0 {
				g.Falling = true
				g.FallingX = float64(boardPaddingPx + slotX*(slotSizePx+slotPaddingPx) + slotSizePx/2)
				g.FallingY = float64(boardPaddingPx + slotSizePx/2)
				g.VelocityY = 0
				g.TargetX = slotX
				g.TargetY = y
				g.FallStopY = float64(boardPaddingPx + (6-y)*(slotSizePx+slotPaddingPx) + slotSizePx/2)
				g.CurrentPlayer = 3 - g.CurrentPlayer
				break
			}
		}
	}

	if g.Falling {
		g.VelocityY += 1
		g.FallingY += g.VelocityY
		if g.FallingY >= g.FallStopY {
			g.Falling = false
			g.Slots[g.TargetX][g.TargetY] = 3 - g.CurrentPlayer
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Draw the slots
	for x := 0; x < 7; x++ {
		for y := 0; y < 6; y++ {
			drawSlot(screen, x, y, g.Slots[x][y])
		}
	}

	if !g.Falling || g.FallingY > boardPaddingPx+slotSizePx*1.5 {
		mx, _ := ebiten.CursorPosition()
		slotX := min(max((mx-boardPaddingPx)/(slotSizePx+slotPaddingPx), 0), 6)
		drawPiece(screen, boardPaddingPx+slotX*(slotSizePx+slotPaddingPx)+slotSizePx/2, boardPaddingPx+slotSizePx/2, g.CurrentPlayer)
	}

	if g.Falling {
		drawPiece(screen, int(g.FallingX), int(g.FallingY), 3-g.CurrentPlayer)
	}
}

func drawSlot(screen *ebiten.Image, x, y int, piece Piece) {
	// Draw the slot
	px := boardPaddingPx + x*(slotSizePx+slotPaddingPx) + slotSizePx/2
	py := boardPaddingPx + (6-y)*(slotSizePx+slotPaddingPx) + slotSizePx/2
	drawPiece(screen, px, py, piece)
}

func drawPiece(screen *ebiten.Image, px, py int, piece Piece) {
	var col color.RGBA
	var stroke bool
	switch piece {
	case Yellow:
		col = color.RGBA{0xff, 0xff, 0x00, 0xff}
	case Red:
		col = color.RGBA{0xff, 0x00, 0x00, 0xff}
	default:
		col = color.RGBA{0xff, 0xff, 0xff, 0xff}
		stroke = true
	}
	if stroke {
		vector.StrokeCircle(screen, float32(px), float32(py), slotSizePx/2, 1, col, true)
	} else {
		vector.DrawFilledCircle(screen, float32(px), float32(py), slotSizePx/2, col, true)
	}
}

const (
	// amount of padding around the board, in pixels
	boardPaddingPx = 40
	// amount of padding around each slot, in pixels
	slotPaddingPx = 20
	// size of each slot, in pixels
	slotSizePx = 100
	// width of the board, in slots
	boardWidth = 7
	// height of the board, in slots
	boardHeight = 6
)

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	screenWidth = boardPaddingPx*2 + boardWidth*(slotSizePx+slotPaddingPx) - slotPaddingPx
	screenHeight = boardPaddingPx*2 + (boardHeight+1)*(slotSizePx+slotPaddingPx) - slotPaddingPx
	return
}
