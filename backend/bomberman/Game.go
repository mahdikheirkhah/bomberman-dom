package bomberman

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const NumberOfRows = 11
const NumberOfColumns = 13
const MaxNumberOfPlayers = 4
const MinNumberOfPlayers = 2
const CellSize = 50

var Colors = []string{"G", "Y", "R", "B"}

type PlayerRespawn struct {
	PlayerIndex int
	RespawnTime time.Time
}

type GameBoard struct {
	Players            []Player                              `json:"players"`
	Bombs              []Bomb                                `json:"bombs"`
	PendingRespawns    []PlayerRespawn                       `json:"-"`
	NumberOfPlayers    int                                   `json:"numberOfPlayers"`
	Panel              [NumberOfRows][NumberOfColumns]string `json:"panel"` // Ex -> Exploade , W -> Wall, D -> Destructible, ""(empty) -> empty cell, B -> Bomb
	CellSize           int                                   `json:"cellSize"`
	Powerups           []Powerup                             `json:"powerups"`
	IsStarted          bool
	ExplodedCells      []ExplodedCellInfo `json:"explodedCells"`
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
	return Colors[g.NumberOfPlayers]
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

// FindInnerCell determines which cell the player is entering based on their pixel position
// Returns the new cell index if crossing a grid boundary, or current cell if not
func (g *GameBoard) FindInnerCell(axis byte, direction byte, location int, playerIndex int) int {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	player := g.Players[playerIndex]
	cellSize := int(g.CellSize)

	// Current grid position
	currentCol := player.Column
	currentRow := player.Row

	switch axis {
	case 'x': // Horizontal movement
		rightBorder := (currentCol + 1) * cellSize
		leftBorder := currentCol * cellSize

		if direction == 'r' { // Moving right
			if location >= rightBorder && currentCol < NumberOfColumns-1 {
				return currentCol + 1
			}
		} else if direction == 'l' { // Moving left
			if location <= leftBorder && currentCol > 0 {
				return currentCol - 1
			}
		}
		return currentCol

	case 'y': // Vertical movement
		bottomBorder := (currentRow + 1) * cellSize
		topBorder := currentRow * cellSize

		if direction == 'd' { // Moving down
			if location >= bottomBorder && currentRow < NumberOfRows-1 {
				return currentRow + 1
			}
		} else if direction == 'u' { // Moving up
			if location <= topBorder && currentRow > 0 {
				return currentRow - 1
			}
		}
		return currentRow
	}

	return 0
}

// FindGridBorderLocation returns the pixel coordinate of the specified grid border
// Includes bounds checking to prevent out-of-range errors
func (g *GameBoard) FindGridBorderLocation(borderName byte, playerIndex int) int {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	player := g.Players[playerIndex]
	cellSize := int(g.CellSize)

	// Clamp row and column to valid ranges
	row := Clamp(player.Row, 0, NumberOfRows-1)
	col := Clamp(player.Column, 0, NumberOfColumns-1)

	switch borderName {
	case 'u': // Top border of current cell
		return row * cellSize
	case 'd': // Bottom border of current cell
		return (row + 1) * cellSize
	case 'l': // Left border of current cell
		return col * cellSize
	case 'r': // Right border of current cell
		return (col + 1) * cellSize
	default:
		return -1 // Invalid border name
	}
}

// Helper function to clamp values between min and max
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (g *GameBoard) FindGridCenterLocation(row, col int) (int, int) {
	x := (col * int(g.CellSize)) + int(g.CellSize/2)
	y := (row * int(g.CellSize)) + int(g.CellSize/2)
	return x, y
}
func (g *GameBoard) RandomStart() {

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
			if row%2 == 1 && col%2 == 1 {
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

	// Print panel to console
	fmt.Println("Game board:")
	for _, line := range g.Panel {
		lineForPrint := ""
		for _, char := range line {
			if char == "" {
				char = "·"
			}
			lineForPrint += char
		}
		fmt.Println(lineForPrint)
	}
}

func (g *GameBoard) SendPlayerAccepted(playerIndex int) {
	msg := map[string]interface{}{
		"type":  "PlayerAccepted",
		"index": playerIndex,
	}

	conn, ok := g.PlayersConnections[playerIndex]
	if !ok {
		log.Printf("error: no connection for player index %d", playerIndex)
		return
	}

	err := conn.WriteJSON(msg)
	if err != nil {
		log.Printf("error writing json to player %d: %v", playerIndex, err)
		conn.Close()
	}
}

func InitGame() *GameBoard {
	g := &GameBoard{
		IsStarted:          false,
		CellSize:           CellSize,
		NumberOfPlayers:    0,
		PlayersConnections: make(map[int]*websocket.Conn),
		BroadcastChannel:   make(chan interface{}, 100),
	}
	g.StartBombWatcher()
	return g
}
func (g *GameBoard) CheckGameEnd() {
	livePlayers := 0
	var lastPlayer Player
	for _, player := range g.Players {
		if player.Lives > 0 {
			livePlayers++
			lastPlayer = player
			log.Printf("Player %s is alive with %d lives\n", player.Name, player.Lives)
		}
	}
	if livePlayers <= 1 && g.IsStarted {
		log.Println("Game end checker started")
		g.IsStarted = false
		msg := map[string]interface{}{
			"type":   "GameState",
			"state":  "GameOver",
			"winner": lastPlayer.Index,
			"player": lastPlayer,
		}
		g.SendMsgToChannel(msg, -1)
		log.Printf("Game over! Winner is player %d\n", lastPlayer.Index)
		go func() {
			time.Sleep(10 * time.Second)
			g.ResetGame()
		}()
	}
}
func (g *GameBoard) ResetGame() {
	g.Mu.Lock()
	g.Players = []Player{}
	g.Bombs = []Bomb{}
	g.Powerups = []Powerup{}
	g.PendingRespawns = []PlayerRespawn{}
	g.NumberOfPlayers = 0
	g.IsStarted = false
	g.ExplodedCells = []ExplodedCellInfo{}
	for conn := range g.PlayersConnections {
		g.PlayersConnections[conn].Close()
	}
	g.CellSize = CellSize
	g.BroadcastChannel = make(chan interface{}, 100)
	g.PlayersConnections = make(map[int]*websocket.Conn)
	g.Panel = [NumberOfRows][NumberOfColumns]string{}
	g.RandomStart()
	LobbyMsg = false
	g.Mu.Unlock()
	go g.StartBombWatcher()
	go g.StartBroadcaster()
	log.Println("Game reset. Waiting for players to join.")
}
