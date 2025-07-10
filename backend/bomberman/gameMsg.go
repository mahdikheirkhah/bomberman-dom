package bomberman

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func (g *GameBoard) StartBroadcaster() {
	for msg := range g.BroadcastChannel {
		g.Mu.Lock()
		conns := make(map[int]*websocket.Conn)
		for k, v := range g.PlayersConnections {
			conns[k] = v
		}
		g.Mu.Unlock()
		for playerIndex, conn := range conns {
			// check for the chat messages sender
			msg := CheckForPlayer(msg, playerIndex)
			err := conn.WriteJSON(msg)
			if err != nil {
				log.Printf("Broadcast error to player %d: %v\n", playerIndex, err)
				conn.Close()
				delete(g.PlayersConnections, playerIndex)
			}
		}
	}
}

func CheckForPlayer(msg interface{}, playerIndex int) interface{} {
	msgMap, ok := msg.(map[string]interface{})
	if !ok {
		return msg
	}

	// Step 2: Extract msgType
	msgType, ok := msgMap["Type"].(string)
	if !ok {
		return msg
	}
	if msgType == "CM" {
		if msgMap["SenderIndex"] == playerIndex {
			msgMap["Filter"] = true
		}
	}
	return msgMap
}
func (g *GameBoard) HandlePlayerMessages(playerIndex int, conn *websocket.Conn) {
	defer func() {
		g.Mu.Lock()
		delete(g.PlayersConnections, playerIndex)
		g.Mu.Unlock()
		conn.Close()
		log.Printf("Connection closed for player %d\n", playerIndex)
	}()

	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading from player %d: %v\n", playerIndex, err)
			break
		}
		// Tag message with player index
		msg["fromPlayer"] = playerIndex
		g.ChooseHandlerForMessages(msg)
	}
}
func (g *GameBoard) SendMsgToChannel(msg any, playerIndex int) {
	select {
	case g.BroadcastChannel <- msg:
		// Message forwarded
	default:
		log.Printf("Broadcast channel full, dropped message from player %d\n", playerIndex)
	}
}
func (g *GameBoard) ChooseHandlerForMessages(msg interface{}) {
	// Step 1: Assert msg is a map[string]interface{}
	msgMap, ok := msg.(map[string]interface{})
	if !ok {
		log.Println("Invalid message format in ChooseHandlerForMessages")
		return
	}
	playerIndex, ok := msgMap["fromPlayer"].(int)
	if !ok {
		log.Println("fromPlayer not found in message")
		return
	}
	// Step 2: Extract msgType
	msgType, ok := msgMap["msgType"].(string)
	if !ok {
		log.Println("msgType not found or not a string")
		return
	}

	// Step 3: Switch based on msgType
	switch msgType {
	case "MS": // Move Start
		direction, ok := msgMap["d"].(string)
		if !ok {
			log.Println("Invalid or missing direction in MS message")
			return
		}
		g.HandleMoveStartMessage(playerIndex, direction)
	case "ME": // Move End
		g.HandleMoveEndMessage(playerIndex)
	case "b": // Bomb
		g.HandleBombMessage(msgMap)
	case "c": // Chat
		g.HandleChatMessage(msgMap)
	case "p": // Powerup (placeholder)
		// Handle powerup logic here
	default:
		log.Println("Unknown msgType:", msgType)
	}
}

type MovePlayerMsg struct {
	MsgType     string `json:"MT"`
	XLocation   int    `json:"XL"`
	YLocation   int    `json:"YL"`
	Direction   string `json:"D"`
	PlayerIndex int    `json:"PI"`
}
type PlantBomb struct {
	MsgType   string `json:"MT"`
	XLocation int    `json:"XL"`
	YLocation int    `json:"YL"`
	Row       int    `json:"R"`
	Column    int    `json:"C"`
}

// type NotMove struct {
// 	MsgType string `json:"MT"`
// }

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
// playerMoveLoop is a goroutine that handles continuous movement for a single player.
func (g *GameBoard) playerMoveLoop(playerIndex int) {
	// Adjust this ticker duration to control movement speed/update frequency
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			g.Mu.Lock() // Acquire lock for this tick's operations
			player := &g.Players[playerIndex]

			// Check conditions to stop movement
			// HasExploaded assumes the mutex is held, which it is here.
			collision := g.FindCollision(playerIndex) // Update collision status
			if !player.IsMoving || player.IsDead || collision == "Ex" || collision == "B" || collision == "D" || collision == "W" {
				// If player is no longer moving, dead, or on an exploded cell, stop this goroutine
				log.Printf("Player %d movement loop stopping due to state change (IsMoving: %t, IsDead: %t, OnCollision: %s).\n",
					playerIndex, player.IsMoving, player.IsDead, collision) // Corrected log message
				player.IsMoving = false // Ensure state is consistent
				if player.StopMoveChan != nil {
					close(player.StopMoveChan) // Close the channel to prevent future starts without recreation
					player.StopMoveChan = nil
				}
				g.Mu.Unlock()
				return // Terminate the goroutine
			}

			// Attempt to move the player
			// g.MovePlayer is assumed to return true on success, false on failure (e.g., hit wall)
			// MovePlayer assumes the mutex is held, which it is here.
			moved := g.MovePlayer(playerIndex, player.DirectionFace) // Corrected field name
			if moved {
				// If player moved successfully, broadcast the new position
				var msg MovePlayerMsg
				msg.MsgType = "M" // General move update
				msg.PlayerIndex = playerIndex
				msg.XLocation = player.XLocation
				msg.YLocation = player.YLocation
				msg.Direction = player.DirectionFace // Corrected field name
				g.SendMsgToChannel(msg, playerIndex)
			} else {
				// If player couldn't move (hit a wall, etc.), stop continuous movement
				log.Printf("Player %d stopped continuous movement due to obstacle.\n", playerIndex)
				player.IsMoving = false
				if player.StopMoveChan != nil {
					close(player.StopMoveChan)
					player.StopMoveChan = nil
				}
				g.Mu.Unlock()
				return // Terminate the goroutine
			}
			g.Mu.Unlock() // Release lock at the end of the tick's operations
		case <-g.Players[playerIndex].StopMoveChan:

			// Received explicit stop signal from HandleMoveEndMessage or disconnect handler
			log.Printf("Player %d movement loop received explicit stop signal.\n", playerIndex)
			g.Mu.Lock() // Acquire lock to ensure state consistency before returning
			player := &g.Players[playerIndex]
			player.IsMoving = false // Ensure state is consistent
			// Channel is already closed by the sender (HandleMoveEndMessage or HandlePlayerMessages defer)
			player.StopMoveChan = nil // Mark as closed
			g.Mu.Unlock()
			return // Terminate the goroutine
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

	g.Mu.Lock()
	defer g.Mu.Unlock()

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
	return
}
