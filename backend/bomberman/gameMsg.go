package bomberman

import (
	"log"

	"github.com/gorilla/websocket"
)

func (g *GameBoard) StartBroadcaster() {
	go func() {
		for msg := range g.BroadcastChannel {
			g.Mu.Lock()
			conns := make(map[int]*websocket.Conn)
			for k, v := range g.PlayersConnections {
				conns[k] = v
			}
			g.Mu.Unlock()
			for playerIndex, conn := range conns {
				err := conn.WriteJSON(msg)
				if err != nil {
					log.Printf("Broadcast error to player %d: %v\n", playerIndex, err)
					conn.Close()
					delete(g.PlayersConnections, playerIndex)
				}
			}

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
		g.ChooseHandlerForMessages(msg)
	}
}
func (g *GameBoard) SendMsgToChannel(msg any, playerIndex int) {
	log.Println("Broadcasting:", msg)
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
		log.Println("Invalid message format")
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
	//move
	case "m":
		if !g.Players[playerIndex].IsDead {
			g.HandleMoveMessage(msgMap)
		}

	//bomb
	case "b":
		if !g.Players[playerIndex].IsDead {
			g.HandleBombMessage(msgMap)
		}

	//chat
	case "c":
		//g.HandleChatMessage(msgMap)

	//power up
	case "p":
	default:
		log.Println("Unknown msgType:", msgType)
	}
}

type MovePlayerMsg struct {
	MsgType   string `json:"MT"`
	XLocation int    `json:"XL"`
	YLocation int    `json:"YL"`
	Row       int    `json:"R"`
	Column    int    `json:"C"`
}
type PlantBomb struct {
	MsgType   string `json:"MT"`
	XLocation int    `json:"XL"`
	YLocation int    `json:"YL"`
	Row       int    `json:"R"`
	Column    int    `json:"C"`
}
type NotMove struct {
	MsgType string `json:"MT"`
}

func (g *GameBoard) HandleMoveMessage(msgMap map[string]interface{}) {
	playerIndex, ok := msgMap["fromPlayer"].(int)
	if !ok {
		log.Println("fromPlayer not found in message")
		return
	}

	direction, ok := msgMap["d"].(string)
	if !ok {
		log.Println("Invalid or missing direction in move message")
		return
	}

	g.Mu.Lock()
	if g.MovePlayer(playerIndex, direction) {
		var msg MovePlayerMsg
		msg.MsgType = "MA" // Move Accepted
		msg.Column = g.Players[playerIndex].Column
		msg.Row = g.Players[playerIndex].Row
		msg.XLocation = g.Players[playerIndex].XLocation
		msg.YLocation = g.Players[playerIndex].YLocation
		g.SendMsgToChannel(msg, playerIndex)
	} else {
		var msg NotMove
		msg.MsgType = "MNA" // Move Not Accpeted
		g.SendMsgToChannel(msg, playerIndex)
	}
	g.Mu.Unlock()
}
func (g *GameBoard) HandleBombMessage(msgMap map[string]interface{}) {
	playerIndex, ok := msgMap["fromPlayer"].(int)
	if !ok {
		log.Println("fromPlayer not found in message")
		return
	}

	g.Mu.Lock()
	bombIndex, err := g.CreateBomb(playerIndex)
	if err != nil {
		var msg PlantBomb
		msg.MsgType = "BA" //Bomb Accepted
		msg.Column = g.Bombs[bombIndex].Column
		msg.Row = g.Bombs[bombIndex].Row
		msg.XLocation = g.Bombs[bombIndex].XLocation
		msg.YLocation = g.Bombs[bombIndex].YLocation
		g.SendMsgToChannel(msg, playerIndex)
	} else {
		var msg NotMove
		msg.MsgType = "BNA" // Bomb Not Accpeted
		g.SendMsgToChannel(msg, playerIndex)
	}
	g.Mu.Unlock()
}
