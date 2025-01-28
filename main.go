package main

import (
	"image"
	"image/color"
	"log"
	"math/rand"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/io/key"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)
var (
    WindowWidth = 1000
    WindowHeight = 250
    GridColumns = 100
    GridRows = GridColumns * WindowHeight / WindowWidth
    FPS = 5.0
    SquareSize = WindowWidth / GridColumns
    BackgroundColor = color.NRGBA{R: 26, G: 26, B: 36, A: 255}
    SquareColor = color.NRGBA{R: 63, G: 78, B: 96, A: 255}
)

func main() {
    go func() {
        w := new(app.Window)
        w.Option(app.Title("Game of Life"))
        w.Option(app.Size(unit.Dp(WindowWidth), unit.Dp(WindowHeight)))
        if err := draw(w); err != nil {
            log.Fatal(err)
        }
        os.Exit(0)
    }()

    app.Main()
}

type Grid struct {
    width  int
    height int
    cells [][]bool
}

func makeGrid(width, height int) Grid {
    grid := Grid{width, height, nil}
    grid.cells = make([][]bool, grid.width)
    for i := range grid.cells {
        grid.cells[i] = make([]bool, grid.height)
    }
    return grid
}

func initRandomGrid(grid *Grid) {
    for i := 0; i < grid.width; i++ {
        for j := 0; j < grid.height; j++ {
            grid.cells[i][j] = rand.Intn(2) == 0
        }
    }
}

func updateGrid(grid *Grid) {
    newGrid := makeGrid(grid.width, grid.height)
    
    for i := 0; i < grid.width; i++ {
        for j := 0; j < grid.height; j++ {
            // count the number of live neighbors
            liveNeighbors := countLiveNeighbors(grid, i, j)

            // apply the rules
            if grid.cells[i][j] {
                newGrid.cells[i][j] = liveNeighbors == 2 || liveNeighbors == 3
            } else {
                newGrid.cells[i][j] = liveNeighbors == 3
            }
        }
    }
    
    *grid = newGrid
}

func countLiveNeighbors(grid *Grid, x, y int) int {
    liveNeighbors := 0
    for i := -1; i <= 1; i++ {
        for j := -1; j <= 1; j++ {
            if i == 0 && j == 0 {
                continue
            }
            
            newX, newY := x+i, y+j
            
            // Wrap around edges (toroidal grid)
            newX = (newX + grid.width) % grid.width
            newY = (newY + grid.height) % grid.height
            
            if grid.cells[newX][newY] {
                liveNeighbors++
            }
        }
    }
    return liveNeighbors
}

func drawSquare(ops *op.Ops, x, y, sizeX, sizeY int, color color.NRGBA) {
    defer clip.Rect(image.Rect(x, y, x+sizeX, y+sizeY)).Push(ops).Pop()
    paint.ColorOp{Color: color}.Add(ops)
    paint.PaintOp{}.Add(ops)
}

func draw(w *app.Window) error {
    var ops op.Ops
    grid := makeGrid(GridColumns, GridRows)
    initRandomGrid(&grid)

    lastUpdateTime := time.Now()
    updateInterval := time.Second / time.Duration(FPS)

    for {
        switch e := w.Event().(type) {
        case app.DestroyEvent:
            return e.Err

        case app.FrameEvent:
            now := time.Now()
            if now.Sub(lastUpdateTime) >= updateInterval {
                updateGrid(&grid)
                lastUpdateTime = now
            }

            gtx := app.NewContext(&ops, e)

            // Handle key events
            for {
                event, ok := gtx.Event(key.Filter{Name: key.NameEscape}, key.Filter{Name: key.NameSpace})
                if !ok {
                    break
                }
                if keyEvent, isKey := event.(key.Event); isKey {
                    switch keyEvent.Name {
                    case key.NameEscape:
                        return nil
                    case key.NameSpace:
                        initRandomGrid(&grid)
                    }
                }
            }

            // Clear the screen
            paint.Fill(gtx.Ops, BackgroundColor)

            // Render the grid
            squareWidth := WindowWidth / GridColumns
            squareHeight := WindowHeight / GridRows
            for i := 0; i < grid.width; i++ {
                for j := 0; j < grid.height; j++ {
                    if grid.cells[i][j] {
                        drawSquare(gtx.Ops, i*squareWidth, j*squareHeight, squareWidth, squareHeight, SquareColor)
                    }
                }
            }

            // Draw the frame
            e.Frame(gtx.Ops)

            // Request another frame
            w.Invalidate()
        }
    }
}