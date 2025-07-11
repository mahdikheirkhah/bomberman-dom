package main

import (
	"backend/bomberman"
	"log"
	"net/http"
)

func main() {
	game := bomberman.InitGame()
	game.RandomStart()
	go game.StartBroadcaster()

	// Bind the /ws route
	http.HandleFunc("/ws", game.HandleWSConnections)

	// Start the server
	log.Println("Server started at :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
