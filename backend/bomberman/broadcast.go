package bomberman

import (
	"encoding/json"
	"io"
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

func (g *GameBoard) HandleWSConnections(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if g.IsStarted {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("you can not join now wait"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Can't read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req JoinRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	g.Mu.Lock()
	err = g.CreatePlayer(req.Name)
	if err != nil {
		g.Mu.Unlock()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(err.Error()))
		return
	}
	playerIndex := g.NumberOfPlayers - 1
	g.Mu.Unlock()

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	g.Mu.Lock()
	g.PlayersConnections[playerIndex] = conn
	g.Mu.Unlock()

	go g.HandlePlayerMessages(playerIndex, conn)
	if g.NumberOfPlayers == MinNumberOfPlayers {

		// Start countdown only once
		once.Do(func() {
			go g.startCountdown()
		})
	}
	if LobbyMsg && !g.IsStarted && g.NumberOfPlayers != MaxNumberOfPlayers {
		stateMsg := StateMsg{
			Type:  "GameState",
			State: "LobbyCountdown",
		}
		g.SendMsgToChannel(stateMsg, -1)
	}

	if g.NumberOfPlayers == MaxNumberOfPlayers {
		g.forceStartGame()
	}
}

func (g *GameBoard) startCountdown() {
	for i := 20; i > 0; i-- {
		msg := map[string]interface{}{
			"type":    "countdown",
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
	g.Mu.Lock()
	defer g.Mu.Unlock()

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
	for i := 10; i > 0; i-- {
		msg := map[string]interface{}{
			"type":    "countdown",
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
