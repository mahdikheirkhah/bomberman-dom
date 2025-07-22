package bomberman
import (
	"log"
	"time"
)

func (g *GameBoard) HandleChatMessage(msgMap map[string]interface{}) {
	var msg Chat
	playerIndex, ok := msgMap["fromPlayer"].(int)
	if !ok {
		log.Println("fromPlayer not found in message")
		return
	}

	Content, ok := msgMap["content"].(string)
	if !ok {
		log.Println("fromPlayer not found in message")
		return
	}
	g.Mu.Lock()
	msg.Type = "CM" // chat message
	msg.Name = g.Players[playerIndex].Name
	msg.Color = g.Players[playerIndex].Color
	msg.SenderIndex = playerIndex
	msg.Content = Content
	msg.Date = time.Now()
	g.SendMsgToChannel(msg, playerIndex)
	g.Mu.Unlock()


}