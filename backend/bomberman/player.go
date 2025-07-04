package bomberman

import (
	"errors"
	"time"
)

const StepSize = 0.5
const MiliBombDelay = 500

type Player struct {
	Name              string        `json:"name"`
	Lives             int           `json:"lives"`
	Score             int           `json:"score"`
	Color             string        `json:"color"`
	Row               int           `json:"row"`
	Column            int           `json:"column"`
	XLocation         float64       `json:"xlocation"`
	YLocation         float64       `json:"yLocation"`
	IsDead            bool          `json:"isDead"`
	NumberOfBombs     int           `json:"numberOfBombs"`
	NumberOfUsedBombs int           `json:"numberOfUsedBombs"`
	BombDelay         time.Duration `json:"bombDelay"`
	StepSize          float64       `json:"stepSize"`
	DirectionFace     byte          `json:"DirectionFace"`
}

func (g *GameBoard) CreatePlayer(name string) error {
	var player Player
	if !g.CanCreateNewPlayer() {
		return errors.New("max number of players of has been reached")
	}

	player.Name = name
	player.Lives = 3
	player.Score = 0
	player.Color = g.FindColor()
	player.Row = g.FindStartRowLocation()
	player.Column = g.FindStartColLocation()
	player.StepSize = StepSize
	player.BombDelay = MiliBombDelay * time.Millisecond
	g.Players = append(g.Players, player)

	g.NumberOfPlayers++

	return nil
}

func (g *GameBoard) MovePlayer(playerIndex int, direction string) bool {
	player := &g.Players[playerIndex]
	var destionation float64
	if player.DirectionFace != byte(direction[0]) {
		player.DirectionFace = byte(direction[0])
	} else {
		switch direction {
		//up
		case "u":
			destionation = player.YLocation + player.StepSize
			player.Row = g.FindInnerCell('y', 'u', destionation, playerIndex)
			if len(g.Panel) <= player.Row || g.Panel[player.Row][player.Column].IsWall || g.Panel[player.Row][player.Column].IsDestructible {
				return false
			}
			player.YLocation = destionation

		//down
		case "d":
			destionation = player.YLocation - player.StepSize
			player.Row = g.FindInnerCell('y', 'd', destionation, playerIndex)
			if player.Row < 0 || g.Panel[player.Row][player.Column].IsWall || g.Panel[player.Row][player.Column].IsDestructible {
				return false
			}
			player.YLocation = destionation

		//rigth
		case "r":
			destionation = player.XLocation + player.StepSize
			player.Column = g.FindInnerCell('x', 'r', destionation, playerIndex)
			if len(g.Panel[0]) <= player.Column || g.Panel[player.Row][player.Column].IsWall || g.Panel[player.Row][player.Column].IsDestructible {
				return false
			}
			player.XLocation = destionation

		//left
		case "l":
			destionation = player.XLocation - player.StepSize
			player.Column = g.FindInnerCell('x', 'r', destionation, playerIndex)
			if player.Column < 0 || g.Panel[player.Row][player.Column].IsWall || g.Panel[player.Row][player.Column].IsDestructible {
				return false
			}
			player.XLocation = destionation
		default:
			return false
		}
	}

	return true
}
