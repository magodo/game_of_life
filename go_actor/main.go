package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime/pprof"
	"time"
)

type CellInfo struct {
	x     int
	y     int
	state bool
}

type Cell struct {
	state     bool
	nextState bool
	x         int
	y         int
	// send out self state, neighbours should listen to it
	selfCh chan bool
	// receive neighbours' state
	neighbourChs []<-chan bool
	// receive controller event
	ctrlEventCh chan interface{}
	// send to controller that self transition is ready
	ctrlReadyCh chan interface{}
	// send to controller that self transition is done, and sync self info
	ctrlCellInfoCh chan CellInfo
}

func (c *Cell) run() {

	// first transit
	c.ctrlReadyCh <- struct{}{}
	<-c.ctrlEventCh
	c.ctrlCellInfoCh <- CellInfo{c.x, c.y, c.state}

	for {
		// broadcast current state update to neighbours
		for i := 0; i < len(c.neighbourChs); i++ {
			c.selfCh <- c.state
		}
		// wait for neighbours' state updates
		var neighbourStates []bool
		for _, ch := range c.neighbourChs {
			neighbourStates = append(neighbourStates, <-ch)
		}
		// resolve next state
		c.nextState = c.resolveNextState(neighbourStates)
		// tell controller ready to transit
		c.ctrlReadyCh <- struct{}{}
		// wait for ctrl transit event
		<-c.ctrlEventCh
		// transit: update current state
		c.state = c.nextState
		// tell controller finish transition
		c.ctrlCellInfoCh <- CellInfo{c.x, c.y, c.state}
	}
}

func (c *Cell) resolveNextState(neighbourStates []bool) bool {
	var aliveNeighbourCount int
	for _, s := range neighbourStates {
		if s {
			aliveNeighbourCount++
		}
	}
	if c.state {
		return aliveNeighbourCount >= 2 && aliveNeighbourCount <= 3
	}
	return aliveNeighbourCount == 3
}

type Controller struct {
	grid       [][]rune
	eventCh    chan interface{}
	readyCh    chan interface{}
	cellInfoCh chan CellInfo
}

func (ctrl *Controller) transit() {
	nCell := len(ctrl.grid) * len(ctrl.grid[0])
	// wait for all cells ready to transit
	for i := 0; i < nCell; i++ {
		<-ctrl.readyCh
	}
	// tell all cells to transit
	for i := 0; i < nCell; i++ {
		ctrl.eventCh <- struct{}{}
	}
	// wait for all cells finish transition and use the received cell info to compose grid
	for i := 0; i < nCell; i++ {
		cellInfo := <-ctrl.cellInfoCh
		indicator := ' '
		if cellInfo.state {
			indicator = '*'
		}
		ctrl.grid[cellInfo.x][cellInfo.y] = indicator
	}
	// dump grid
	output := ""
	for x := range ctrl.grid {
		for y := range ctrl.grid[x] {
			output += " " + string(ctrl.grid[x][y])
		}
		output += "\n"
	}
	print("\033[H\033[2J")
	fmt.Println(output)
}

func launchCells(row int, col int, ctrl *Controller) {
	// initialize cells
	tmpGrid := make([][]*Cell, row)
	for i := range tmpGrid {
		tmpGrid[i] = make([]*Cell, col)
	}
	for x := 0; x < row; x++ {
		for y := 0; y < col; y++ {
			neighbourCount := 0
			for nx := x - 1; nx <= x+1; nx++ {
				for ny := y - 1; ny <= y+1; ny++ {
					if nx == x && ny == y {
						continue
					}
					if nx < 0 || nx >= row {
						continue
					}
					if ny < 0 || ny >= col {
						continue
					}
					neighbourCount++
				}
			}
			tmpGrid[x][y] = &Cell{
				state:          rand.Intn(2) == 0,
				x:              x,
				y:              y,
				selfCh:         make(chan bool, neighbourCount),
				ctrlEventCh:    ctrl.eventCh,
				ctrlReadyCh:    ctrl.readyCh,
				ctrlCellInfoCh: ctrl.cellInfoCh,
			}
		}
	}
	// loop again to update neighbours' channel
	for x := 0; x < row; x++ {
		for y := 0; y < col; y++ {
			for nx := x - 1; nx <= x+1; nx++ {
				for ny := y - 1; ny <= y+1; ny++ {
					if nx == x && ny == y {
						continue
					}
					if nx < 0 || nx >= row {
						continue
					}
					if ny < 0 || ny >= col {
						continue
					}
					tmpGrid[x][y].neighbourChs = append(tmpGrid[x][y].neighbourChs, tmpGrid[nx][ny].selfCh)
				}
			}
			// launch
			go tmpGrid[x][y].run()
		}
	}
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {

	// cli parser
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalln(err)
		}
		defer pprof.StopCPUProfile()
	}

	// signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	row, col, interval := 150, 200, time.Millisecond*20
	grid := make([][]rune, row)
	for i := range grid {
		grid[i] = make([]rune, col)
	}
	ctrl := &Controller{
		grid:       grid,
		eventCh:    make(chan interface{}, row*col),
		readyCh:    make(chan interface{}, row*col),
		cellInfoCh: make(chan CellInfo, row*col),
	}

	launchCells(row, col, ctrl)

loop:
	for {
		select {
		case <-sigCh:
			break loop
		default:
			after := time.After(interval)
			start := time.Now()
			ctrl.transit()
			fmt.Printf("Duration: %v\n", time.Now().Sub(start))
			<-after
		}
	}
}
