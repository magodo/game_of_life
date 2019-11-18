package main

import (
	"fmt"
	"time"
)

var nextStateFuncMap = map[int]nextStateFunc{
	0: aliveNextState,
	1: deadNextState,
}

type nextStateFunc func(aliveNeibours int) bool

func aliveNextState(aliveNeibours int) bool {
	return aliveNeibours >= 2 && aliveNeibours <= 3
}

func deadNextState(aliveNeibours int) bool {
	return aliveNeibours == 3
}

type Cell struct {
	state int
	x     int
	y     int
	world [][]Cell
}

func (c Cell) relativeNeibour(horizontal, vertical int) (Cell, bool) {
	neibourX := c.x + horizontal
	neibourY := c.y + vertical
	if neibourX < 0 || neibourX >= len(c.world) {
		return Cell{}, false
	}
	if neibourY < 0 || neibourY >= len(c.world[0]) {
		return Cell{}, false
	}
	return c.world[neibourX][neibourY], true
}

func (c Cell) nextState(aliveNeibours int) int {
	isAlive := nextStateFuncMap[c.state](aliveNeibours)
	if isAlive {
		return 1
	}
	return 0
}

func (c Cell) NewCell(newWorld [][]Cell) Cell {
	aliveNeibours := 0
	for rx := -1; rx <= 1; rx++ {
		for ry := -1; ry <= 1; ry++ {
			if rx == 0 && ry == 0 {
				continue
			}
			if neibour, ok := c.relativeNeibour(rx, ry); ok {
				aliveNeibours += neibour.state
			}
		}
	}
	return Cell{c.nextState(aliveNeibours), c.x, c.y, newWorld}
}

func NewWorld() [][]Cell {
	N := 3
	rect := make([][]Cell, N)
	for i := 0; i < N; i++ {
		rect[i] = make([]Cell, N)
		for j := 0; j < N; j++ {
			rect[i][j] = Cell{1, i, j, rect}
		}
	}
	return rect
}

func Refresh(world [][]Cell) (newWorld [][]Cell) {
	newWorld = NewWorld()
	for row := 0; row < len(world); row++ {
		for col := 0; col < len(world[0]); col++ {
			newWorld[row][col] = world[row][col].NewCell(newWorld)
		}
	}
	return newWorld
}

func PrintWorld(world [][]Cell) {
	for x := 0; x < len(world); x++ {
		for y := 0; y < len(world[0]); y++ {
			fmt.Printf("%d  ", world[x][y].state)
		}
		fmt.Println("")
	}
	fmt.Println("")
}

func main() {
	world := NewWorld()
	for {
		PrintWorld(world)
		world = Refresh(world)
		time.Sleep(time.Second)
	}
}
