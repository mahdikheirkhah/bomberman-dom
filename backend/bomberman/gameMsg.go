package bomberman

import (
	"log"

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
