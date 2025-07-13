package bomberman

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		// Allow any localhost origin (with or without port)
		if strings.HasPrefix(origin, "http://localhost") {
			return true
		}
		return false
	},
}

type JoinRequest struct {
	Name string `json:"name"`
}

type StateMsg struct {
	Type  string `json:"type"`
	State string `json:"state"`
}

var once sync.Once
var LobbyMsg bool

var lobbyCountdownTimer = 5 // 20 for production
var startCountdownTimer = 3 // 10 for production

func (g *GameBoard) HandleWSConnections(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling new WS connection")

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		// The upgrader writes a response on error, so we just return.
		return
	}

	if g.IsStarted {
		log.Println("Game already started, closing connection")
		conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(1008, "Game is full"),
			time.Now().Add(time.Second))
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		log.Println("Player name is missing from query parameters")
		// http.Error(w, "Player name is required as a query parameter", http.StatusBadRequest)
		conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(1008, "Player name is missing from query parameters"),
			time.Now().Add(time.Second))
		return

	}

	g.Mu.Lock()
	err = g.CreatePlayer(name)
	if err != nil {
		g.Mu.Unlock()
		log.Println("Error creating player:", err)
		conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(1008, "Error creating player"),
			time.Now().Add(time.Second))
		return
	}
	playerIndex := g.NumberOfPlayers - 1

	g.PlayersConnections[playerIndex] = conn
	log.Printf("Player %s connected successfully as player %d\n", name, playerIndex)
	g.Mu.Unlock()

	stateMsg := StateMsg{
		Type:  "GameState",
		State: "PlayerAccepted",
	}
	g.SendMsgToChannel(stateMsg, playerIndex)

	// Send the current list of players to all clients
	playerListMsg := map[string]interface{}{
		"type":    "player_list",
		"players": g.Players,
	}
	g.SendMsgToChannel(playerListMsg, -1) // -1 sends to all

	go g.HandlePlayerMessages(playerIndex, conn)

	if g.NumberOfPlayers == MinNumberOfPlayers {
		log.Println("Minimum number of players reached. Starting countdown.")
		// Start countdown only once
		once.Do(func() {
			go g.startCountdown()
		})
	}

	if g.NumberOfPlayers == MaxNumberOfPlayers {
		log.Println("Maximum number of players reached. Forcing game start.")
		g.forceStartGame()
	}
}

func (g *GameBoard) startCountdown() {
	stateMsg := StateMsg{
		Type:  "GameState",
		State: "LobbyCountdown",
	}
	g.SendMsgToChannel(stateMsg, -1)

	for i := lobbyCountdownTimer; i > 0; i-- {
		msg := map[string]interface{}{
			"type":    "lobbyCountdown",
			"seconds": i,
		}
		g.SendMsgToChannel(msg, -1)
		time.Sleep(1 * time.Second)
		LobbyMsg = true
		g.Mu.Lock()
		if g.NumberOfPlayers == MaxNumberOfPlayers {
			g.Mu.Unlock()
			return // Game already started
		}
		g.Mu.Unlock()
	}

	g.forceStartGame()
}

func (g *GameBoard) forceStartGame() {

	if g.IsStarted {
		return
	}
	g.IsStarted = true

	stateMsg := StateMsg{
		Type:  "GameState",
		State: "GameCountdown",
	}
	g.SendMsgToChannel(stateMsg, -1)

	// 10 seconds to start
	for i := startCountdownTimer; i > 0; i-- {
		msg := map[string]interface{}{
			"type":    "gameCountdown",
			"seconds": i,
		}
		g.SendMsgToChannel(msg, -1)
		time.Sleep(1 * time.Second)
	}

	stateMsg = StateMsg{
		Type:  "GameState",
		State: "GameStarted",
	}
	g.SendMsgToChannel(stateMsg, -1)
	msg := struct {
		Type            string                                `json:"type"`
		Players         []Player                              `json:"players"`
		NumberOfPlayers int                                   `json:"numberOfPlayers"`
		Panel           [NumberOfRows][NumberOfColumns]string `json:"panel"`
	}{
		Type:            "gameStart",
		Players:         g.Players,
		NumberOfPlayers: g.NumberOfPlayers,
		Panel:           g.Panel,
	}

	g.SendMsgToChannel(msg, -1)
}
