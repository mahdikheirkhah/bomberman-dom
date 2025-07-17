package bomberman

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections
	},
}

type JoinRequest struct {
	Name string `json:"name"`
}

type StateMsg struct {
	Type  string `json:"type"`
	State string `json:"state"`
}

var LobbyMsg bool


// CheckNameHandler handles HTTP requests to check if a player name is already taken or if the game is started.
// It expects a 'name' query parameter.
// Responds with JSON: {"isTaken": true/false, "reason": "..."}
func (g *GameBoard) CheckNameHandler(w http.ResponseWriter, r *http.Request) {
	// CORS headers: Allow requests from your frontend origin
	// In a production environment, replace "http://localhost:8000" with your actual frontend domain.
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS") // Allow GET and preflight OPTIONS
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type") // Allow Content-Type header

	// Handle preflight OPTIONS requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	log.Println("CheckNameHandler: Received request to check name availability.")

	name := r.URL.Query().Get("name")
	name = strings.TrimSpace(name)
	if name == "" {
		log.Println("CheckNameHandler: Missing 'name' query parameter.")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Name parameter is required."})
		return
	}

	g.Mu.Lock() // Acquire a read lock to check IsStarted and IsPlayerNameTaken
	isStarted := g.IsStarted
	UUID, err := g.CreatePlayer(name)

	g.Mu.Unlock() // Release the read lock immediately after reading state

	// Prepare the response
	response := make(map[string]interface{}) // Use interface{} to allow mixed types

	if isStarted {
		response["reason"] = "game_already_started"
		w.WriteHeader(http.StatusConflict) // 409 Conflict
		log.Printf("CheckNameHandler: Game already started. Name '%s' cannot join.", name)
	} else if err != nil {
		response["reason"] = err.Error()
		w.WriteHeader(http.StatusConflict) // 409 Conflict
		log.Printf("CheckNameHandler: Name '%s' cannot join: %s", name, err.Error())
	} else {
		response["uuid"] = UUID
		w.WriteHeader(http.StatusOK) // 200 OK
		log.Printf("CheckNameHandler: Name '%s' is available.", name)
	}

	// Encode the response to JSON and send it
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("CheckNameHandler: Error encoding JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (g *GameBoard) HandleWSConnections(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling new WS connection")

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		// The upgrader writes a response on error, so we just return.
		return
	}

	UUID := r.URL.Query().Get("UUID")
	if UUID == "" {
		log.Println("Player UUID is missing from query parameters")
		// http.Error(w, "Player name is required as a query parameter", http.StatusBadRequest)
		conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(1008, "Player UUID is missing from query parameters"),
			time.Now().Add(time.Second))
		return

	}

	g.Mu.Lock()
	playerIndex := g.GetPlayerByUUID(UUID)
	if playerIndex == -1 {
		g.Mu.Unlock()
		errMsg := fmt.Sprintf("Error finding player with UUID %s", UUID)
		log.Println(errMsg)
		conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(1008, errMsg),
			time.Now().Add(time.Second))
		return
	}

	g.PlayersConnections[playerIndex] = conn
	log.Printf("Player %s connected successfully as player %d\n", g.Players[playerIndex].Name, playerIndex)
	g.Mu.Unlock()

	g.SendPlayerAccepted(playerIndex)

	// Send the current list of players to all clients
	playerListMsg := map[string]interface{}{
		"type":    "player_list",
		"players": g.Players,
	}
	g.SendMsgToChannel(playerListMsg, -1) // -1 sends to all

	go g.HandlePlayerMessages(playerIndex, conn)
	var once sync.Once
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
