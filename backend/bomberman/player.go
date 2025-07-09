package bomberman

import (
	"errors"
	"time"
)

const StepSize = 5
const MiliBombDelay = 500
const BombRange = 2

type Player struct {
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
	player.XLocation, player.YLocation = g.FindGridCenterLocation(player.Row, player.Column)
	player.StepSize = StepSize
	player.BombDelay = MiliBombDelay * time.Millisecond
	player.BombRange = BombRange
	g.Players = append(g.Players, player)

	g.NumberOfPlayers++

	return nil
}

func (g *GameBoard) MovePlayer(playerIndex int, direction string) bool {
	player := &g.Players[playerIndex]
	step := player.StepSize

	if player.DirectionFace != byte(direction[0]) {
		player.DirectionFace = byte(direction[0])
	} else {
		switch direction {
		case "u":
			dest := player.YLocation - step
			newRow := g.FindInnerCell('y', 'u', dest, playerIndex)
			if newRow < 0 || g.Panel[newRow][player.Column] == "W" || g.Panel[newRow][player.Column] == "D" || g.Panel[newRow][player.Column] == "B" {
				return false
			}
			player.Row = newRow
			player.YLocation = dest

		case "d":
			dest := player.YLocation + step
			newRow := g.FindInnerCell('y', 'd', dest, playerIndex)
			if newRow >= NumberOfRows || g.Panel[newRow][player.Column] == "W" || g.Panel[newRow][player.Column] == "D" || g.Panel[newRow][player.Column] == "B" {
				return false
			}
			player.Row = newRow
			player.YLocation = dest

		case "r":
			dest := player.XLocation + step
			newCol := g.FindInnerCell('x', 'r', dest, playerIndex)
			if newCol >= NumberOfColumns || g.Panel[player.Row][newCol] == "W" || g.Panel[player.Row][newCol] == "D" || g.Panel[player.Row][newCol] == "B" {
				return false
			}
			player.Column = newCol
			player.XLocation = dest

		case "l":
			dest := player.XLocation - step
			newCol := g.FindInnerCell('x', 'l', dest, playerIndex)
			if newCol < 0 || g.Panel[player.Row][newCol] == "W" || g.Panel[player.Row][newCol] == "D" || g.Panel[player.Row][newCol] == "B" {
				return false
			}
			player.Column = newCol
			player.XLocation = dest
		default:
			return false
		}
	}
	return true
}
