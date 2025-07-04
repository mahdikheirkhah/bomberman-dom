package bomberman

import (
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const NumberOfRows = 11
const NumberOfColumns = 13
const MaxNumberOfPlayers = 4
const MinNumberOfPlayers = 2
const CellSize = 64

var Colors = []string{"G", "Y", "R", "B"}

type GameBoard struct {
	Players            []Player                              `json:"players"`
	Bombs              []Bomb                                `json:"bombs"`
	NumberOfPlayers    int                                   `json:"numberOfPlayers"`
	Panel              [NumberOfRows][NumberOfColumns]string `json:"panel"` // Ex -> Exploade , W -> Wall, D -> Destructible, ""(empty) -> empty cell
	CellSize           int                                   `json:"cellSize"`
	IsStarted          bool
	PlayersConnections map[int]*websocket.Conn

	BroadcastChannel chan interface{}

	Mu sync.Mutex
}

type GameCell struct {
	IsOccupied     bool `json:"isOccupied"`
	IsWall         bool `json:"isWall"`
	IsDestructible bool `json:"isDestructible"`
	IsExploaded    bool `json:"isExploaded"`
}

func (g *GameBoard) CanCreateNewPlayer() bool {
	if 0 < g.NumberOfPlayers+1 && g.NumberOfPlayers+1 <= MaxNumberOfPlayers {
		return true
	}
	return false
}

func (g *GameBoard) FindColor() string {
	return Colors[g.NumberOfPlayers+1]
}

func (g *GameBoard) FindStartRowLocation() int {
	if g.NumberOfPlayers+1 == 1 || g.NumberOfPlayers+1 == 2 {
		return 0
	}
	return NumberOfRows - 1
}

func (g *GameBoard) FindStartColLocation() int {
	if g.NumberOfPlayers+1 == 1 || g.NumberOfPlayers+1 == 3 {
		return 0
	}
	return NumberOfColumns - 1
}

func (g *GameBoard) HasExploaded(row, col int) bool {
	return g.Panel[row][col] == "Ex"
}

func (g *GameBoard) FindInnerCell(axis byte, direction byte, location int, playerIndex int) int {
	col := g.Players[playerIndex].Column
	row := g.Players[playerIndex].Row

	if axis == 'x' {
		if direction == 'r' && location >= g.FindGridBorderLocation('r', playerIndex) {
			return col + 1
		}
		if direction == 'l' && location <= g.FindGridBorderLocation('l', playerIndex) {
			return col - 1
		}
		return col
	} else if axis == 'y' {
		if direction == 'u' && location <= g.FindGridBorderLocation('u', playerIndex) {
			return row - 1
		}
		if direction == 'd' && location >= g.FindGridBorderLocation('d', playerIndex) {
			return row + 1
		}
		return row
	}

	return 0
}

func (g *GameBoard) FindGridBorderLocation(borderName byte, playerIndex int) int {
	row := g.Players[playerIndex].Row
	col := g.Players[playerIndex].Column
	cellSize := int(g.CellSize)

	switch borderName {
	case 'u':
		return row * cellSize // top border
	case 'd':
		return (row + 1) * cellSize // bottom border
	case 'l':
		return col * cellSize
	case 'r':
		return (col + 1) * cellSize
	}
	return -1
}

func (g *GameBoard) FindGridCenterLocation(row, col int) (int, int) {
	x := (col * int(g.CellSize)) + int(g.CellSize/2)
	y := (row * int(g.CellSize)) + int(g.CellSize/2)
	return x, y
}
func (g *GameBoard) RandomStart() {
	rand.Seed(time.Now().UnixNano())

	// Define safe zones around each player start (row, col) + adjacent cells
	safeZones := map[int][][2]int{
		0: {{0, 0}, {0, 1}, {1, 0}},                                                                                                    // Top-left
		1: {{0, NumberOfColumns - 1}, {0, NumberOfColumns - 2}, {1, NumberOfColumns - 1}},                                              // Top-right
		2: {{NumberOfRows - 1, 0}, {NumberOfRows - 2, 0}, {NumberOfRows - 1, 1}},                                                       // Bottom-left
		3: {{NumberOfRows - 1, NumberOfColumns - 1}, {NumberOfRows - 2, NumberOfColumns - 1}, {NumberOfRows - 1, NumberOfColumns - 2}}, // Bottom-right
	}

	// Step 1: Fill grid with walls
	for row := 0; row < NumberOfRows; row++ {
		for col := 0; col < NumberOfColumns; col++ {
			var cell string

			// Step 1.1: Place indestructible wall at even-even positions
			if row%2 == 0 && col%2 == 0 {
				cell = "W"
			} else {
				// Step 1.2: Randomly place destructible walls (30% chance)
				if rand.Float64() < 0.3 {
					cell = "D"
				}
			}

			g.Panel[row][col] = cell
		}
	}

	// Step 2: Clear player spawn zones
	for i := 0; i < MaxNumberOfPlayers; i++ {
		for _, pos := range safeZones[i] {
			row, col := pos[0], pos[1]
			g.Panel[row][col] = "" // empty cell
		}
	}
}

func InitGame() *GameBoard {
	g := &GameBoard{
		IsStarted:          false,
		CellSize:           CellSize,
		NumberOfPlayers:    0,
		PlayersConnections: make(map[int]*websocket.Conn),
		BroadcastChannel:   make(chan interface{}),
	}
	return g
}
