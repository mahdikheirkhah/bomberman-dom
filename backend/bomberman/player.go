package bomberman

import (
	"errors"
	"time"

	"github.com/google/uuid"
)



func (g *GameBoard) CreatePlayer(name string) (string, error) {
	var player Player
	if !g.CanCreateNewPlayer() {
		return "", errors.New("max number of players of has been reached")
	}
	if len(name) > 14 {
		return "", errors.New("name should be less than 15 characters")
	}
	for _, p := range g.Players {
		if p.Name == name {
			return "", errors.New("player name is already taken")
		}
	}
	player.UUID = uuid.New().String()
	player.Index = g.NumberOfPlayers
	player.Name = name
	player.Lives = 3 
	player.Score = 0
	player.Color = g.FindColor()
	player.Row = g.FindStartRowLocation()
	player.Column = g.FindStartColLocation()
	player.InitialRow = player.Row
	player.InitialColumn = player.Column
	player.XLocation, player.YLocation = player.Column*int(g.CellSize), player.Row*int(g.CellSize)
	player.StepSize = StepSize
	player.BombDelay = BombDelay * time.Second
	player.BombRange = BombRange
	player.NumberOfBombs = 3
	player.NumberOfUsedBombs = 0
	player.IsDead = false
	player.IsHurt = false
	g.Players = append(g.Players, player)

	g.NumberOfPlayers++

	return player.UUID, nil
}

func (g *GameBoard) GetPlayerByUUID(UUID string) int {
	for i, p := range g.Players {
		if p.UUID == UUID {
			return i
		}
	}
	return -1
}
