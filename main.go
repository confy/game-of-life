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
    WINDOW_WIDTH = 1000
    WINDOW_HEIGHT = 1000
    GRID_SIZE = 100
    FPS = 5.0
    SQUARE_SIZE = WINDOW_HEIGHT/GRID_SIZE
    BACKGROUND_COLOR = color.NRGBA{R: 26, G: 26, B: 36, A: 255}
    SQUARE_COLOR = color.NRGBA{R: 63, G: 78, B: 96, A: 255}
)

func main() {
    go func() {
        w := new(app.Window)
        w.Option(app.Title("Game of Life"))
        w.Option(app.Size(unit.Dp(WINDOW_WIDTH), unit.Dp(WINDOW_HEIGHT)))
        if err := draw(w); err != nil {
            log.Fatal(err)
        }
        os.Exit(0)
    }()

    app.Main()
}

type Grid struct {
    size  int
    cells [][]bool
}

func makeGrid(size int) Grid {
    grid := Grid{size: size}
    grid.cells = make([][]bool, grid.size)
    for i := range grid.cells {
        grid.cells[i] = make([]bool, grid.size)
    }
    return grid
}

func initRandomGrid(grid *Grid) {
    for i := 0; i < grid.size; i++ {
        for j := 0; j < grid.size; j++ {
            grid.cells[i][j] = rand.Intn(2) == 0
        }
    }
}

func updateGrid(grid *Grid) {
    newGrid := makeGrid(grid.size)
    
    for i := 0; i < grid.size; i++ {
        for j := 0; j < grid.size; j++ {
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
            newX = (newX + grid.size) % grid.size
            newY = (newY + grid.size) % grid.size
            
            if grid.cells[newX][newY] {
                liveNeighbors++
            }
        }
    }
    return liveNeighbors
}

func drawSquare(ops *op.Ops, startX, startY, size int, color color.NRGBA) {
    defer clip.Rect(image.Rect(startX, startY, startX+size, startY+size)).Push(ops).Pop()
    paint.ColorOp{Color: color}.Add(ops)
    paint.PaintOp{}.Add(ops)
}

func draw(w *app.Window) error {
    var ops op.Ops
    grid := makeGrid(GRID_SIZE)
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

            for {
				event, ok := gtx.Event(key.Filter{Name: key.NameEscape}, key.Filter{Name: key.NameSpace})
				if !ok {
					break
				}
				switch event := event.(type) {
				case key.Event:
					if event.Name == key.NameEscape {
						return nil
					}
                    if event.Name == key.NameSpace {
                        initRandomGrid(&grid)
                    }
				}
			}
            // Clear the screen
            paint.Fill(gtx.Ops, BACKGROUND_COLOR)

            // Render the grid
            for i := 0; i < grid.size; i++ {
                for j := 0; j < grid.size; j++ {
                    if grid.cells[i][j] {
                        drawSquare(gtx.Ops, i*SQUARE_SIZE, j*SQUARE_SIZE, SQUARE_SIZE, SQUARE_COLOR)
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