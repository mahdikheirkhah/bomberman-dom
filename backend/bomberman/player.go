package bomberman

import "errors"

type Player struct {
	Name   string `json:"name"`
	Lives  int    `json:"lives"`
	Score  int    `json:"score"`
	Color  string `json:"color"`
	Row    int    `json:"row"`
	Column int    `json:"column"`
}

func CreatePlayer(name string, gameboard *GameBoard) (Player, error) {
	var player Player
	if !gameboard.CanCreateNewPlayer() {
		return Player{}, errors.New("Max number of players of has been reached!")
	}

	player.Name = name
	player.Lives = 3
	player.Score = 0
	player.Color = gameboard.FindColor()
	player.Row = gameboard.FindStartRowLocation()
	player.Column = gameboard.FindStartColLocation()

	return player, nil
}
