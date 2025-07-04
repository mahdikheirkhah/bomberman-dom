package main

import (
	"backend/bomberman"
	"fmt"
)

func main() {
	game := bomberman.InitGame()
	fmt.Println(game.Panel)
}
