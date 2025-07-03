package bomberman

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

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

func (g *GameBoard) HandleWSConnections(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Can't read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse JSON into struct
	var req JoinRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	g.Mu.Lock()
	err = g.CreatePlayer(req.Name)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(err.Error()))
		return
	}
	g.Mu.Unlock()

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	g.Mu.Lock()
	g.PlayersConnections[g.NumberOfPlayers-1] = conn
	msg := struct {
		Players         []Player                                `json:"players"`
		NumberOfPlayers int                                     `json:"numberOfPlayers"`
		Panel           [NumberOfRows][NumberOfColumns]GameCell `json:"panel"`
	}{
		Players:         g.Players,
		NumberOfPlayers: g.NumberOfPlayers,
		Panel:           g.Panel,
	}

	g.Mu.Unlock()

	err = conn.WriteJSON(msg)
	if err != nil {
		fmt.Println("error at pong:", err)
	}
}
