package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	NumberOfRows       = 20
	NumberOfColumns    = 20
	MaxNumberOfPlayers = 4
	MinNumberOfPlayers = 2
)

type GameCell struct{}

type Player struct {
	ID                string        `json:"id"`
	Name              string        `json:"name"`
	Lives             int           `json:"lives"`
	Score             int           `json:"score"`
	Color             string        `json:"color"`
	Row               int           `json:"row"`
	Column            int           `json:"column"`
	IsDead            bool          `json:"isDead"`
	NumberOfBombs     int           `json:"numberOfBombs"`
	NumberOfUsedBombs int           `json:"numberOfUsedBombs"`
	BombDelay         time.Duration `json:"bombDelay"`
}

type Bomb struct{}

type GameBoard struct {
	Players         []Player                                `json:"players"`
	Bombs           []Bomb                                  `json:"bombs"`
	NumberOfPlayers int                                     `json:"numberOfPlayers"`
	Panel           [NumberOfRows][NumberOfColumns]GameCell `json:"panel"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var game = struct {
	mu      sync.Mutex
	board   GameBoard
	clients map[*websocket.Conn]string
}{
	board: GameBoard{
		Players: make([]Player, 0, MaxNumberOfPlayers),
	},
	clients: make(map[*websocket.Conn]string),
}

func main() {
	fs := http.FileServer(http.Dir("../frontend"))
	http.Handle("/", fs)
	http.HandleFunc("/api/join", corsMiddleware(handleJoin))
	http.HandleFunc("/ws", handleWebSocket)

	fmt.Println("Fake Bomberman server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	}
}

func handleJoin(w http.ResponseWriter, r *http.Request) {
	log.Println("[handleJoin] Received new join request")
	var player Player
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		log.Println("[handleJoin] Error: Invalid request body")
		return
	}

	game.mu.Lock()

	if len(game.board.Players) >= MaxNumberOfPlayers {
		game.mu.Unlock()
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Game is full"})
		log.Println("[handleJoin] Error: Game is full")
		return
	}

	player = Player{
		ID:   uuid.New().String(),
		Name: req.Name,
	}

	game.board.Players = append(game.board.Players, player)
	game.board.NumberOfPlayers++
	log.Printf("[handleJoin] Player %s added to game", player.Name)

	game.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"playerId": player.ID})

	broadcastGameStatus()
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	playerId := r.URL.Query().Get("playerId")
	if playerId == "" {
		log.Println("Player ID is missing")
		return
	}

	game.mu.Lock()
	game.clients[conn] = playerId
	game.mu.Unlock()

	log.Printf("Player %s connected via WebSocket", playerId)
	broadcastGameStatus()

	if len(game.board.Players) == 1 {
		go simulatePlayerJoins()
	}

	for {
		// Keep the connection alive
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}
	}

	game.mu.Lock()
	delete(game.clients, conn)
	game.mu.Unlock()
	log.Printf("Player %s disconnected", playerId)
}

func broadcastGameStatus() {
	game.mu.Lock()

	playerNames := make([]string, len(game.board.Players))
	for i, p := range game.board.Players {
		playerNames[i] = p.Name
	}

	status := "waiting_for_players"
	if len(game.board.Players) >= MinNumberOfPlayers {
		status = "ready_to_start"
	}

	message := map[string]interface{}{
		"type": "game_status",
		"payload": map[string]interface{}{
			"status":  status,
			"players": playerNames,
		},
	}

	b, _ := json.Marshal(message)

	clients := make([]*websocket.Conn, 0, len(game.clients))
	for client := range game.clients {
		clients = append(clients, client)
	}

	game.mu.Unlock()

	for _, client := range clients {
		if err := client.WriteMessage(websocket.TextMessage, b); err != nil {
			log.Println("Error broadcasting game status:", err)
		}
	}
}

func simulatePlayerJoins() {
	for i := 2; i <= MaxNumberOfPlayers; i++ {
		time.Sleep(time.Duration(rand.Intn(3)+2) * time.Second)

		game.mu.Lock()
		if len(game.board.Players) >= MaxNumberOfPlayers {
			game.mu.Unlock()
			break
		}

		player := Player{
			ID:   uuid.New().String(),
			Name: fmt.Sprintf("Player %d", i),
		}
		game.board.Players = append(game.board.Players, player)
		log.Println("Player added:" + player.Name)
		game.board.NumberOfPlayers++
		game.mu.Unlock()

		broadcastGameStatus()
	}

	time.Sleep(3 * time.Second)
	broadcastGameStart()
}

func broadcastGameStart() {
	game.mu.Lock()
	defer game.mu.Unlock()

	message := map[string]interface{}{
		"type": "game_start",
	}

	b, _ := json.Marshal(message)

	for client := range game.clients {
		if err := client.WriteMessage(websocket.TextMessage, b); err != nil {
			log.Println("Error broadcasting game start:", err)
		}
	}
}
