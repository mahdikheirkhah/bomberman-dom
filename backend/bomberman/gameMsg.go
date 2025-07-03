package bomberman

import (
	"log"

	"github.com/gorilla/websocket"
)

func (g *GameBoard) StartBroadcaster() {
	go func() {
		for msg := range g.BroadcastChannel {
			g.Mu.Lock()
			for playerIndex, conn := range g.PlayersConnections {
				err := conn.WriteJSON(msg)
				if err != nil {
					log.Printf("Broadcast error to player %d: %v\n", playerIndex, err)
					conn.Close()
					delete(g.PlayersConnections, playerIndex)
				}
			}
			g.Mu.Unlock()
		}
	}()
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
		var msg map[string]interface{} // or define a proper struct if you know the schema
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading from player %d: %v\n", playerIndex, err)
			break
		}
		// Tag message with player index
		msg["fromPlayer"] = playerIndex

		select {
		case g.BroadcastChannel <- msg:
			// Message forwarded
		default:
			log.Printf("Broadcast channel full, dropped message from player %d\n", playerIndex)
		}
	}
}

// func (g *GameBoard) ChooseHandlerForMessages(msg interface{}) {
// 	// Step 1: Assert msg is a map[string]interface{}
// 	msgMap, ok := msg.(map[string]interface{})
// 	if !ok {
// 		log.Println("Invalid message format")
// 		return
// 	}

// 	// Step 2: Extract msgType
// 	msgType, ok := msgMap["msgType"].(string)
// 	if !ok {
// 		log.Println("msgType not found or not a string")
// 		return
// 	}

// 	// Step 3: Switch based on msgType
// 	switch msgType {
// 	//move
// 	case "m":
// 		g.HandleMoveMessage(msgMap)

// 	//bomb
// 	case "b":
// 		g.HandleBombMessage(msgMap)

// 	//chat
// 	case "c":
// 		g.HandleChatMessage(msgMap)

// 	//power up
// 	case "p":
// 	default:
// 		log.Println("Unknown msgType:", msgType)
// 	}
// }
