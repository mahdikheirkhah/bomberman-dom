package bomberman

import (
	"errors"
	"time"
)

const StepSize = 5
const MiliBombDelay = 500
const BombRange = 2

type Player struct {
	Index             int           `json:"index"`
	Name              string        `json:"name"`
	Lives             int           `json:"lives"`
	Score             int           `json:"score"`
	Color             string        `json:"color"`
	Row               int           `json:"row"`
	Column            int           `json:"column"`
	XLocation         int           `json:"xlocation"`
	YLocation         int           `json:"yLocation"`
	IsDead            bool          `json:"isDead"`
	NumberOfBombs     int           `json:"numberOfBombs"`
	NumberOfUsedBombs int           `json:"numberOfUsedBombs"`
	BombDelay         time.Duration `json:"bombDelay"`
	BombRange         int           `json:"bombRange"`
	StepSize          int           `json:"stepSize"`
	DirectionFace     string        `json:"DirectionFace"`
	IsMoving          bool          `json:"isMoving"`
	StopMoveChan      chan struct{} `json:"-"` // Channel to signal the player's movement goroutine to stop
}

const PlayerSize = 48

func (g *GameBoard) CreatePlayer(name string) error {
	var player Player
	if !g.CanCreateNewPlayer() {
		return errors.New("max number of players of has been reached")
	}
	player.Index = g.NumberOfPlayers
	player.Name = name
	player.Lives = 3
	player.Score = 0
	player.Color = g.FindColor()
	player.Row = g.FindStartRowLocation()
	player.Column = g.FindStartColLocation()
	player.XLocation, player.YLocation = player.Column*int(g.CellSize), player.Row*int(g.CellSize)
	player.StepSize = StepSize
	player.BombDelay = MiliBombDelay * time.Millisecond
	player.BombRange = BombRange
	player.NumberOfBombs = 3
	player.NumberOfUsedBombs = 0
	player.IsDead = false
	g.Players = append(g.Players, player)

	g.NumberOfPlayers++

	return nil
}
