package bomberman

import (
	"log"
	"time"
)

type MovePlayerMsg struct {
	MsgType     string `json:"MT"`
	XLocation   int    `json:"XL"`
	YLocation   int    `json:"YL"`
	Direction   string `json:"D"`
	PlayerIndex int    `json:"PI"`
}

// HandleMoveStartMessage initiates continuous movement for a player.
func (g *GameBoard) HandleMoveStartMessage(playerIndex int, direction string) {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	player := &g.Players[playerIndex]
	if player.IsDead {
		log.Printf("Player %d is dead, cannot start movement.\n", playerIndex)
		return
	}
	collision := g.FindCollision(playerIndex)
	if collision == "Ex" || collision == "B" || collision == "D" || collision == "W" { // Check if player is on an exploded cell
		log.Printf("Player %d is on an exploded cell, cannot start movement.\n", playerIndex)
		return
	}

	player.IsMoving = true
	player.DirectionFace = direction

	// Start a new movement goroutine only if one isn't already running
	// Check if the channel is nil or if it was previously closed and needs recreation
	if player.StopMoveChan == nil {
		player.StopMoveChan = make(chan struct{}) // Create a new stop channel
		go g.playerMoveLoop(playerIndex)          // Start the continuous movement loop
		log.Printf("Player %d started continuous movement in direction: %s\n", playerIndex, direction)
	} else {
		log.Printf("Player %d already has a movement loop running.\n", playerIndex)
	}
}

// HandleMoveEndMessage stops continuous movement for a player.
func (g *GameBoard) HandleMoveEndMessage(playerIndex int) {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	player := &g.Players[playerIndex]
	if player.IsMoving {
		player.IsMoving = false
		// Signal the movement goroutine to stop
		if player.StopMoveChan != nil {
			close(player.StopMoveChan)
			player.StopMoveChan = nil // Mark as closed
			log.Printf("Player %d stopped continuous movement.\n", playerIndex)
		}
	}
}

// playerMoveLoop is a goroutine that handles continuous movement for a single player.
// It checks the player's movement state and updates their position at regular intervals.
func (g *GameBoard) playerMoveLoop(playerIndex int) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			g.Mu.Lock() // Hold lock for entire operation

			player := &g.Players[playerIndex]
			if !player.IsMoving || player.IsDead {
				g.Mu.Unlock()
				return
			}

			// Check collision before moving
			if collision := g.FindCollision(playerIndex); collision == "Ex" || collision == "B" || collision == "D" || collision == "W" {
				player.IsMoving = false
				if player.StopMoveChan != nil {
					close(player.StopMoveChan)
					player.StopMoveChan = nil
				}
				g.Mu.Unlock()
				return
			}

			// Perform movement
			moved := g.MovePlayer(playerIndex, player.DirectionFace)
			if moved {
				// Broadcast new position
				msg := MovePlayerMsg{
					MsgType:     "M",
					PlayerIndex: playerIndex,
					XLocation:   player.XLocation,
					YLocation:   player.YLocation,
					Direction:   player.DirectionFace,
				}
				g.SendMsgToChannel(msg, playerIndex)
			} else {
				player.IsMoving = false
				if player.StopMoveChan != nil {
					close(player.StopMoveChan)
					player.StopMoveChan = nil
				}
			}

			g.Mu.Unlock()

		case <-g.Players[playerIndex].StopMoveChan:
			g.Mu.Lock()
			player := &g.Players[playerIndex]
			player.IsMoving = false
			if player.StopMoveChan != nil {
				player.StopMoveChan = nil
			}
			g.Mu.Unlock()
			return
		}
	}
}

// HandleMoveMessage (original, now superseded by MS/ME) - Kept for reference or if still used for single moves
// This function is now only called if msgType is "M" (single move), not "MS" or "ME".
// If you only want continuous movement, you might remove this or adapt it.
func (g *GameBoard) HandleMoveMessage(msgMap map[string]interface{}) {
	playerIndex, ok := msgMap["fromPlayer"].(int)
	if !ok {
		log.Println("fromPlayer not found in message")
		return
	}

	direction, ok := msgMap["D"].(string) // Changed from "d" to "D" for consistency with MS/ME
	if !ok {
		log.Println("Invalid or missing direction in move message")
		return
	}

	// g.Mu.Lock()
	// defer g.Mu.Unlock()

	player := &g.Players[playerIndex]
	if player.IsDead || g.HasExploaded(player.Row, player.Column) {
		log.Printf("Player %d is dead or on exploded cell, cannot move.\n", playerIndex)
		return
	}

	if g.MovePlayer(playerIndex, direction) { // Assuming MovePlayer updates player's X/Y
		g.SendMoveMsg(playerIndex)
	}
}

func (g *GameBoard) SendMoveMsg(playerIndex int) {
	var msg MovePlayerMsg
	msg.MsgType = "M" // General move update
	msg.PlayerIndex = playerIndex
	msg.XLocation = g.Players[playerIndex].XLocation
	msg.YLocation = g.Players[playerIndex].YLocation
	msg.Direction = g.Players[playerIndex].DirectionFace
	g.SendMsgToChannel(msg, playerIndex)
}

func (g *GameBoard) FindCollision(playerIndex int) string {
	player := g.Players[playerIndex]
	cellSize := int(g.CellSize)

	// Get current cell
	currentRow := player.YLocation / cellSize
	currentCol := player.XLocation / cellSize

	// Check current cell first
	if currentRow >= 0 && currentRow < NumberOfRows &&
		currentCol >= 0 && currentCol < NumberOfColumns {
		if g.Panel[currentRow][currentCol] != "" {
			return g.Panel[currentRow][currentCol]
		}
	}

	// Check adjacent cells based on direction and position within current cell
	xInCell := player.XLocation % cellSize
	yInCell := player.YLocation % cellSize

	// Check right border
	if xInCell > cellSize-5 && currentCol < NumberOfColumns-1 {
		if g.Panel[currentRow][currentCol+1] != "" {
			return g.Panel[currentRow][currentCol+1]
		}
	}

	// Check left border
	if xInCell < 5 && currentCol > 0 {
		if g.Panel[currentRow][currentCol-1] != "" {
			return g.Panel[currentRow][currentCol-1]
		}
	}

	// Check bottom border
	if yInCell > cellSize-5 && currentRow < NumberOfRows-1 {
		if g.Panel[currentRow+1][currentCol] != "" {
			return g.Panel[currentRow+1][currentCol]
		}
	}

	// Check top border
	if yInCell < 5 && currentRow > 0 {
		if g.Panel[currentRow-1][currentCol] != "" {
			return g.Panel[currentRow-1][currentCol]
		}
	}

	return ""
}
func (g *GameBoard) FindDistanceToBorder(playerIndex int, borderName string) int {
	row := g.Players[playerIndex].Row
	col := g.Players[playerIndex].Column
	cellSize := int(g.CellSize)
	player := &g.Players[playerIndex]
	switch borderName {
	case "u":
		distance := player.YLocation - (row * cellSize)
		return distance
	case "d":
		distance := (row)*cellSize - player.YLocation
		return distance
	case "l":
		distance := player.XLocation - (col * cellSize)
		return distance
	case "r":
		distance := (col)*cellSize - player.XLocation
		return distance
	}
	return -1
}

func (g *GameBoard) MovePlayer(playerIndex int, direction string) bool {
	player := &g.Players[playerIndex]
	step := player.StepSize
	cellSize := int(g.CellSize)

	originalX := player.XLocation
	originalY := player.YLocation

	switch direction {
	case "u":
		player.YLocation -= step
		if player.YLocation < 0 {
			player.YLocation = 0
		}
	case "d":
		player.YLocation += step
		if player.YLocation >= NumberOfRows*cellSize {
			player.YLocation = NumberOfRows*cellSize - 1
		}
	case "l":
		player.XLocation -= step
		if player.XLocation < 0 {
			player.XLocation = 0
		}
	case "r":
		player.XLocation += step
		if player.XLocation >= NumberOfColumns*cellSize {
			player.XLocation = NumberOfColumns*cellSize - 1
		}
	}

	// Update row/column
	player.Row = player.YLocation / cellSize
	player.Column = player.XLocation / cellSize

	// Check if we hit something
	if collision := g.FindCollision(playerIndex); collision != "" {
		// Revert position
		player.XLocation = originalX
		player.YLocation = originalY
		player.Row = originalY / cellSize
		player.Column = originalX / cellSize
		return false
	}

	return true
}
