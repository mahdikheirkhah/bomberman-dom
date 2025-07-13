package bomberman

import (
	"errors"
	"log"
	"time"
)

// BombExplosionDuration defines how long cells stay 'Ex' after an explosion before clearing.
const BombExplosionDuration = 1 * time.Second

// Position represents a row and column in the game grid.
type Position struct {
	Row        int  `json:"row"`
	Col        int  `json:"col"`
	CellOnFire bool `json:"CellOnFire"`
}

// Bomb represents a bomb placed on the game board.
type Bomb struct {
	Row                 int       `json:"row"`
	Column              int       `json:"column"`
	XLocation           int       `json:"xlocation"`     // Assuming these are for rendering
	YLocation           int       `json:"yLocation"`     // Assuming these are for rendering
	ExplosionTime       time.Time `json:"explosionTime"` // When the bomb will explode
	OwnPlayerIndex      int       `json:"playerIndex"`   // ID of the player who placed this bomb
	InitialIntersection bool      `json:"initialIntersection"`
}

// Powerup represents an item that can be collected by players.
type Powerup struct {
	Type     string `json:"type"`
	Value    int    `json:"value"`
	Row      int    `json:"row"`
	Column   int    `json:"column"`
	IsHidden bool   `json:"isHidden"`
}

// ExplodedCellInfo tracks a cell that has been exploded and its scheduled clear time.
type ExplodedCellInfo struct {
	Position  Position
	ClearTime time.Time // When this cell should revert from "Ex" to ""
}
type ExploadeCellsMsg struct {
	MsgType   string     `json:"MT"`
	Positions []Position `json:"positions"`
	BombRow   int        `json:"bombRow"`
	BombCol   int        `json:"bombCol"`
}

type PlayerExplosionMsg struct {
	Type        string `json:"type"`
	Lives       int    `json:"lives"`
	Color       string `json:"color"`
	PlayerIndex int    `json:"playerIndex"`
}
type PLayerDeath struct {
	Type   string `json:"type"`
	Player Player `json:"player"`
}

type PlantBomb struct {
	MsgType     string `json:"MT"`
	XLocation   int    `json:"XL"`
	YLocation   int    `json:"YL"`
	Row         int    `json:"R"`
	Column      int    `json:"C"`
	PlayerIndex int    `json:"PI"`
}

func (g *GameBoard) HandleBombMessage(msgMap map[string]interface{}) {
	playerIndex, ok := msgMap["fromPlayer"].(int)
	if !ok {
		log.Println("fromPlayer not found in message")
		return
	}
	g.Mu.Lock()
	defer g.Mu.Unlock()
	bombIndex, err := g.CreateBomb(playerIndex)
	if err != nil {
		log.Println("Error creating bomb:", err)
		return
	}
	var msg PlantBomb
	msg.MsgType = "BA" //Bomb Accepted
	msg.Column = g.Bombs[bombIndex].Column
	msg.Row = g.Bombs[bombIndex].Row
	msg.XLocation = g.Bombs[bombIndex].XLocation
	msg.YLocation = g.Bombs[bombIndex].YLocation
	msg.PlayerIndex = playerIndex
	g.SendMsgToChannel(msg, playerIndex)
}

func (g *GameBoard) RespawnPlayer(playerIndex int) {
	g.Players[playerIndex].Row = g.Players[playerIndex].InitialRow
	g.Players[playerIndex].Column = g.Players[playerIndex].InitialColumn
	g.Players[playerIndex].XLocation = g.Players[playerIndex].Column * g.CellSize
	g.Players[playerIndex].YLocation = g.Players[playerIndex].Row * g.CellSize
}

// CheckExplosion iterates through players and reduces lives if they are on an "Ex" cell.
func (g *GameBoard) CheckExplosion() {
	for i := range g.Players {
		if g.Players[i].IsDead {
			continue
		}

		collision := g.FindCollision(i)
		if collision == "Ex" {
			g.Players[i].Lives--

			if g.Players[i].IsMoving {
				g.Players[i].IsMoving = false
				if g.Players[i].StopMoveChan != nil {
					close(g.Players[i].StopMoveChan)
					g.Players[i].StopMoveChan = nil // Mark as closed
				}
			}

			if g.Players[i].Lives <= 0 {
				g.PlayerDeath(i)
			} else {
				// Player was hit but is not dead, add to respawn queue.
				g.PendingRespawns = append(g.PendingRespawns, PlayerRespawn{
					PlayerIndex: i,
					RespawnTime: time.Now().Add(BombExplosionDuration),
				})

				msg := PlayerExplosionMsg{
					Type:        "PLD", // player live decrease
					Lives:       g.Players[i].Lives,
					Color:       g.Players[i].Color,
					PlayerIndex: i,
				}
				g.SendMsgToChannel(msg, -1)
			}
		}
	}
}

func (g *GameBoard) PlayerDeath(playerIndex int) {
	g.NumberOfPlayers--
	g.Players[playerIndex].IsDead = true

	msg := PLayerDeath{
		Type:   "PD", // player dead
		Player: g.Players[playerIndex],
	}
	g.SendMsgToChannel(msg, -1)
}

// HasExploaded checks if a given position on the game panel is currently marked as "Ex".
func (g *GameBoard) HasExploaded(row, col int) bool {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	if row < 0 || row >= len(g.Panel) || col < 0 || col >= len(g.Panel[0]) {
		return false // Position is out of bounds
	}
	return g.Panel[row][col] == "Ex"
}

// CanCreateBomb checks if a player is allowed to place another bomb.
func (g *GameBoard) CanCreateBomb(playerIndex int) bool {

	if playerIndex < 0 || playerIndex >= len(g.Players) {
		return false // Invalid player index
	}
	return g.Players[playerIndex].NumberOfUsedBombs < g.Players[playerIndex].NumberOfBombs
}

// CreateBomb creates a new bomb at the player's current position.
func (g *GameBoard) CreateBomb(playerIndex int) (int, error) {

	if playerIndex < 0 || playerIndex >= len(g.Players) {
		return -1, errors.New("invalid player index")
	}

	if !g.CanCreateBomb(playerIndex) {
		return -1, errors.New("can not create a new bomb: player has reached bomb limit")
	}

	// Increment the count of bombs used by this player
	g.Players[playerIndex].NumberOfUsedBombs++

	var bomb Bomb
	bomb.Column = g.Players[playerIndex].Column
	bomb.Row = g.Players[playerIndex].Row
	bomb.XLocation = bomb.Column * g.CellSize
	bomb.YLocation = bomb.Row * g.CellSize
	bomb.ExplosionTime = time.Now().Add(g.Players[playerIndex].BombDelay)
	bomb.OwnPlayerIndex = playerIndex // Associate bomb with the player who placed it
	bomb.InitialIntersection = true

	g.Bombs = append(g.Bombs, bomb)
	bombIndex := len(g.Bombs) - 1
	return bombIndex, nil
}

// CalculateBombRange determines all grid positions that will be affected by a bomb explosion.
// It does NOT modify the game board; it only calculates the positions.
func (g *GameBoard) CalculateBombRange(bombRow, bombCol, bombRange int) []Position {
	var affectedPositions []Position

	// Always include the bomb's own position
	affectedPositions = append(affectedPositions, Position{Row: bombRow, Col: bombCol})

	// Check explosion range upwards
	for row := bombRow - 1; row >= 0 && bombRow-row <= bombRange; row-- {
		if g.Panel[row][bombCol] == "W" { // Stop if a wall is encountered
			break
		}
		affectedPositions = append(affectedPositions, Position{Row: row, Col: bombCol})
		if g.Panel[row][bombCol] == "D" { // Stop after destroying a destructible block
			break
		}
	}

	// Check explosion range downwards
	for row := bombRow + 1; row < len(g.Panel) && row-bombRow <= bombRange; row++ {
		if g.Panel[row][bombCol] == "W" {
			break
		}
		affectedPositions = append(affectedPositions, Position{Row: row, Col: bombCol})
		if g.Panel[row][bombCol] == "D" {
			break
		}
	}

	// Check explosion range to the left
	for col := bombCol - 1; col >= 0 && bombCol-col <= bombRange; col-- {
		if g.Panel[bombRow][col] == "W" {
			break
		}
		affectedPositions = append(affectedPositions, Position{Row: bombRow, Col: col})
		if g.Panel[bombRow][col] == "D" {
			break
		}
	}

	// Check explosion range to the right
	for col := bombCol + 1; col < len(g.Panel[0]) && col-bombCol <= bombRange; col++ {
		if g.Panel[bombRow][col] == "W" {
			break
		}
		affectedPositions = append(affectedPositions, Position{Row: bombRow, Col: col})
		if g.Panel[bombRow][col] == "D" {
			break
		}
	}

	return affectedPositions
}

// ApplyExplosion marks cells as 'Ex' or "" (if destructible) and schedules them for clearing.
// It also decrements the bomb count for the player who placed it.
func (g *GameBoard) ApplyExplosion(bomb Bomb) {

	// Find the player who placed this bomb to get their current bomb range
	var player *Player
	for i := range g.Players {
		if i == bomb.OwnPlayerIndex {
			player = &g.Players[i]
			break
		}
	}
	if player == nil {
		// Log an error: Player not found, perhaps disconnected or an invalid ID.
		// This bomb's explosion won't decrement a player's bomb count.
		log.Printf("Player with index %d not found for bomb explosion.", bomb.OwnPlayerIndex)
		return
	}

	// Calculate the actual positions affected by this bomb's explosion
	affectedPositions := g.CalculateBombRange(bomb.Row, bomb.Column, player.BombRange)

	// Check for chain reactions with other bombs
	for i := range g.Bombs {
		// Skip the bomb that is currently exploding
		if g.Bombs[i].Row == bomb.Row && g.Bombs[i].Column == bomb.Column {
			continue
		}

		for _, pos := range affectedPositions {
			if g.Bombs[i].Row == pos.Row && g.Bombs[i].Column == pos.Col {
				// This bomb is in the blast radius. Trigger it to explode almost immediately.
				g.Bombs[i].ExplosionTime = time.Now()
				break // Move to the next bomb in g.Bombs once a match is found
			}
		}
	}
	var msg ExploadeCellsMsg
	msg.MsgType = "EXC" // Exploaded Cells
	msg.BombRow = bomb.Row
	msg.BombCol = bomb.Column
	for _, pos := range affectedPositions {
		// Ensure the position is within the game board boundaries
		if pos.Row < 0 || pos.Row >= len(g.Panel) || pos.Col < 0 || pos.Col >= len(g.Panel[0]) {
			continue // Skip out-of-bounds positions
		}

		cell := &g.Panel[pos.Row][pos.Col]
		if *cell != "W" { // If it's not a solid wall, it should explode. This includes "D" blocks.
			// TODO: If *cell == "D", consider spawning a powerup here.
			*cell = "Ex" // Mark as exploded
			g.ExplodedCells = append(g.ExplodedCells, ExplodedCellInfo{Position: pos, ClearTime: time.Now().Add(BombExplosionDuration)})
			msg.Positions = append(msg.Positions, Position{Row: pos.Row, Col: pos.Col, CellOnFire: true})
		}
	}
	g.SendMsgToChannel(msg, -1)

	// Decrement the number of bombs currently used by the player
	player.NumberOfUsedBombs--
	// Check if any players were caught in the explosion and update their lives/status
	g.CheckExplosion()
}

// ClearExpiredExplosions iterates through the list of exploded cells and
// reverts them to empty ("") if their clear time has passed.
func (g *GameBoard) ClearExpiredExplosions() {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	var remainingExplodedCells []ExplodedCellInfo
	now := time.Now()
	var msg ExploadeCellsMsg
	for _, info := range g.ExplodedCells {
		if now.After(info.ClearTime) {
			// Time to clear this cell
			// Ensure position is still within bounds before modifying the panel

			if info.Position.Row >= 0 && info.Position.Row < len(g.Panel) &&
				info.Position.Col >= 0 && info.Position.Col < len(g.Panel[0]) {
				// Only clear the cell if it's still marked as "Ex".
				// This prevents clearing a cell that has been re-exploded by another bomb.
				if g.Panel[info.Position.Row][info.Position.Col] == "Ex" {
					g.Panel[info.Position.Row][info.Position.Col] = ""
										msg.MsgType = "OF" // Turn Off Fire
					msg.Positions = append(msg.Positions, Position{Row: info.Position.Row, Col: info.Position.Col})
				}
			}
		} else {
			// This cell has not yet expired, keep it in the list
			remainingExplodedCells = append(remainingExplodedCells, info)
		}
	}
	if msg.MsgType == "OF" {
		g.SendMsgToChannel(msg, -1)
	}
	g.ExplodedCells = remainingExplodedCells // Update the list with only unexpired cells
}

func (g *GameBoard) ProcessRespawns() {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	var remainingRespawns []PlayerRespawn
	now := time.Now()

	for _, respawn := range g.PendingRespawns {
		if now.After(respawn.RespawnTime) {
			g.RespawnPlayer(respawn.PlayerIndex)
			player := g.Players[respawn.PlayerIndex]
			msg := struct {
				Type        string `json:"type"`
				PlayerIndex int    `json:"playerIndex"`
				XLocation   int    `json:"xlocation"`
				YLocation   int    `json:"yLocation"`
			}{
				Type:        "PR", // Player Respawn
				PlayerIndex: respawn.PlayerIndex,
				XLocation:   player.XLocation,
				YLocation:   player.YLocation,
			}
			g.SendMsgToChannel(msg, -1)
		} else {
			remainingRespawns = append(remainingRespawns, respawn)
		}
	}
	g.PendingRespawns = remainingRespawns
}

// StartBombWatcher starts a goroutine that periodically checks for bomb explosions
// and clears expired exploded cells.
func (g *GameBoard) StartBombWatcher() {
	go func() {
		// Check every 100 milliseconds for bombs to explode and cells to clear
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop() // Ensure the ticker is stopped when the goroutine exits

		for range ticker.C {
			g.checkBombs()             // Process bombs that are due to explode
			g.ClearExpiredExplosions() // Clear cells whose explosion effect has timed out
			g.ProcessRespawns()        // Process pending respawns
		}
	}()
}

// checkBombs processes bombs that have reached their explosion time.
func (g *GameBoard) checkBombs() {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	var remainingBombs []Bomb
	now := time.Now()

	for _, bomb := range g.Bombs {
		if now.After(bomb.ExplosionTime) {
			// This bomb is ready to explode
			g.ApplyExplosion(bomb)
			// This bomb is now "used up" and will not be added back to remainingBombs
		} else {
			// This bomb has not yet exploded, keep it for the next check
			remainingBombs = append(remainingBombs, bomb)
		}
	}
	g.Bombs = remainingBombs // Update the list of active bombs
}

// Note: The original BombRowRange, BombColRange, and BombExploisonTime constants
// are no longer directly used in the explosion logic if Player.BombRange and
// BombExplosionDuration are the source of truth. You might consider removing them
// if they are truly redundant to avoid confusion.
