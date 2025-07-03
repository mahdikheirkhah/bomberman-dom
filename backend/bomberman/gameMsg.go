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

		// Send to broadcast channel
		g.BroadcastChannel <- msg
	}
}
