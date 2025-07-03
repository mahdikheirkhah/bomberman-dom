package main

import (
	"backend/bomberman"
	"fmt"
	"log"
	"net/http"
)

func main() {
	game := bomberman.InitGame()
	fmt.Println(game.Panel)
	// Bind the /ws route
	http.HandleFunc("/ws", game.HandleWSConnections)

	// Start the server
	log.Println("Server started at :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
