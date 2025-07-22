package bomberman

import (
	"log"
	"time"
)



// HandleMoveStartMessage initiates continuous movement for a player.
func (g *GameBoard) HandleMoveStartMessage(playerIndex int, direction string) {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	if !g.IsStarted {
		return
	}

	player := &g.Players[playerIndex]
	if player.IsDead {
		log.Printf("Player %d is dead, cannot start movement.\n", playerIndex)
		return
	}

	if player.IsHurt {
		log.Printf("Player %d is hurt, cannot start movement.\n", playerIndex)
		return
	}

	// NEW LOGIC: Only prevent movement if the player is currently on a *non-traversable* collision type.
	// "Ex" (exploded cells) should be traversable.
	collision := g.FindCollision(playerIndex)
	if collision == "B" || collision == "D" || collision == "W" || collision == "P" { // Add "P" for player collision if applicable here
		log.Printf("Player %d cannot start movement due to collision with: %s\n", playerIndex, collision)
		return
	}

	player.IsMoving = true
	player.DirectionFace = direction

	// Start a new movement goroutine only if one isn't already running
	if player.StopMoveChan == nil {
		player.StopMoveChan = make(chan struct{}) // Create a new stop channel
		go g.playerMoveLoop(playerIndex)          // Start the continuous movement loop
		//log.Printf("Player %d started continuous movement in direction: %s\n", playerIndex, direction)
	} else {
		//log.Printf("Player %d already has a movement loop running.\n", playerIndex)
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
			//log.Printf("Player %d stopped continuous movement.\n", playerIndex)
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

			if !g.IsStarted {
				g.Mu.Unlock()
				return
			}

			player := &g.Players[playerIndex]
			if !player.IsMoving || player.IsDead || player.IsHurt{
				g.Mu.Unlock()
				return
			}

			// NEW LOGIC: Check collision *after* attempting a move.
			// The current player's cell can be "Ex", but they should still be able to move *from* it.
			// The actual blocking logic should primarily be within MovePlayer.

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
				// If MovePlayer returns false, it means the player hit an impassable object.
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

	// g.Mu.Lock() // Consider if a lock is needed here if it's a single, non-looping move
	// defer g.Mu.Unlock()

	player := &g.Players[playerIndex]
	if player.IsDead { // Removed the HasExploaded check as "Ex" cells should be traversable.
		log.Printf("Player %d is dead, cannot move.\n", playerIndex)
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

	// Player's bounding box corners
	topLeftX, topLeftY := player.XLocation, player.YLocation
	topRightX, topRightY := player.XLocation+PlayerSize-1, player.YLocation
	bottomLeftX, bottomLeftY := player.XLocation, player.YLocation+PlayerSize-1
	bottomRightX, bottomRightY := player.XLocation+PlayerSize-1, player.YLocation+PlayerSize-1

	// Convert corner coordinates to grid cells
	topLeftRow, topLeftCol := topLeftY/cellSize, topLeftX/cellSize
	topRightRow, topRightCol := topRightY/cellSize, topRightX/cellSize
	bottomLeftRow, bottomLeftCol := bottomLeftY/cellSize, bottomLeftX/cellSize
	bottomRightRow, bottomRightCol := bottomRightY/cellSize, bottomRightX/cellSize

	// Use a map to check unique cells to avoid redundant checks
	cellsToCheck := map[[2]int]bool{
		{topLeftRow, topLeftCol}:         true,
		{topRightRow, topRightCol}:       true,
		{bottomLeftRow, bottomLeftCol}:   true,
		{bottomRightRow, bottomRightCol}: true,
	}

	for cell := range cellsToCheck {
		row, col := cell[0], cell[1]
		if row >= 0 && row < NumberOfRows && col >= 0 && col < NumberOfColumns {
			cellContent := g.Panel[row][col]
			powerupIndex := g.FindPowerupAt(row, col)
			if cellContent == "" && powerupIndex != -1 {
				g.EatPowerup(playerIndex, powerupIndex)
			}
			// FindCollision should report *all* collisions, including "Ex".
			// It's up to the *caller* of FindCollision to decide what to do with "Ex".
			if cellContent != "" { // Report any non-empty cell content
				return cellContent
			}
		}
	}

	// Check for bomb collisions
	for _, bomb := range g.Bombs {
		// Only consider active bombs that are not the player's own initial bomb placement
		if bomb.OwnPlayerIndex == playerIndex && bomb.InitialIntersection {
			continue // Player can initially pass through their own bomb
		}

		if player.XLocation < bomb.XLocation+cellSize &&
			player.XLocation+PlayerSize > bomb.XLocation &&
			player.YLocation < bomb.YLocation+cellSize &&
			player.YLocation+PlayerSize > bomb.YLocation {
			return "B" // Collision with a bomb
		}
	}

	// Check for player collisions
	for i, otherPlayer := range g.Players {
		if i == playerIndex || otherPlayer.IsDead {
			continue
		}

		// Calculate collision for actual player bounding boxes, not just cell centers
		if player.XLocation < otherPlayer.XLocation+PlayerSize &&
			player.XLocation+PlayerSize > otherPlayer.XLocation &&
			player.YLocation < otherPlayer.YLocation+PlayerSize &&
			player.YLocation+PlayerSize > otherPlayer.YLocation {
			return "P" // Collision with another player
		}
	}

	return "" // No collision
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

	// Temporarily move the player to the new potential position
	switch direction {
	case "u":
		player.YLocation -= step
		if player.YLocation < 0 {
			player.YLocation = 0
		}
	case "d":
		player.YLocation += step
		if player.YLocation+PlayerSize > NumberOfRows*cellSize {
			player.YLocation = NumberOfRows*cellSize - PlayerSize
		}
	case "l":
		player.XLocation -= step
		if player.XLocation < 0 {
			player.XLocation = 0
		}
	case "r":
		player.XLocation += step
		if player.XLocation+PlayerSize > NumberOfColumns*cellSize {
			player.XLocation = NumberOfColumns*cellSize - PlayerSize
		}
	}

	// Update row/column based on the player's center for current *potential* position
	player.Row = (player.YLocation + PlayerSize/2) / cellSize
	player.Column = (player.XLocation + PlayerSize/2) / cellSize

	// Check if we hit an *impassable* object after the tentative move
	if collision := g.FindCollision(playerIndex); collision != "" && collision != "Ex" { // "Ex" cells are ignored here, which is correct.
		// If collision occurs with an impassable object (W, B, D, P), attempt snapping
		// If snapping doesn't resolve it, revert position.

		// Snapping logic to align with grid if close enough, and re-check collision
		movedBySnap := false
		if direction == "l" || direction == "r" { // Moving horizontally, check vertical alignment
			verticalOffset := player.YLocation % cellSize
			if verticalOffset <= movementTolerance {
				player.YLocation -= verticalOffset
				movedBySnap = true
			} else if cellSize-verticalOffset <= movementTolerance {
				player.YLocation += cellSize - verticalOffset
				movedBySnap = true
			}
		} else if direction == "u" || direction == "d" { // Moving vertically, check horizontal alignment
			horizontalOffset := player.XLocation % cellSize
			if horizontalOffset <= movementTolerance {
				player.XLocation -= horizontalOffset
				movedBySnap = true
			} else if cellSize-horizontalOffset <= movementTolerance {
				player.XLocation += cellSize - horizontalOffset
				movedBySnap = true
			}
		}

		// After potential snap, re-check for collision with impassable objects
		if newCollision := g.FindCollision(playerIndex); newCollision != "" && newCollision != "Ex" {
			// Still colliding. Try 1px move from original position.
			player.XLocation = originalX
			player.YLocation = originalY

			step = 1
			switch direction {
			case "u":
				player.YLocation -= step
			case "d":
				player.YLocation += step
			case "l":
				player.XLocation -= step
			case "r":
				player.XLocation += step
			}

			// Clamp for 1px move
			if player.YLocation < 0 {
				player.YLocation = 0
			} else if player.YLocation+PlayerSize > NumberOfRows*cellSize {
				player.YLocation = NumberOfRows*cellSize - PlayerSize
			}
			if player.XLocation < 0 {
				player.XLocation = 0
			} else if player.XLocation+PlayerSize > NumberOfColumns*cellSize {
				player.XLocation = NumberOfColumns*cellSize - PlayerSize
			}

			// Final collision check for 1px move
			if finalCollision := g.FindCollision(playerIndex); finalCollision != "" && finalCollision != "Ex" {
				player.XLocation = originalX
				player.YLocation = originalY
				return false
			}
		} else if !movedBySnap && (collision == "W" || collision == "B" || collision == "D" || collision == "P") {
			// If not snapped and collided with a wall/bomb/destructible/player, revert.
			// This handles cases where player is directly hitting a wall without being "off-center".
			player.XLocation = originalX
			player.YLocation = originalY
			return false
		}
		// If collision was with an "Ex" cell (which is allowed) or snapping resolved it, continue.
	}

	// Update bomb intersection status (this logic is fine)
	for i := range g.Bombs {
		bomb := &g.Bombs[i]
		if bomb.OwnPlayerIndex == playerIndex && bomb.InitialIntersection {
			if player.XLocation >= bomb.XLocation+cellSize ||
				player.XLocation+PlayerSize <= bomb.XLocation ||
				player.YLocation >= bomb.YLocation+cellSize ||
				player.YLocation+PlayerSize <= bomb.YLocation {
				// Player has moved off the bomb
				bomb.InitialIntersection = false
			}
		}
	}

	return true // Movement successful
}
