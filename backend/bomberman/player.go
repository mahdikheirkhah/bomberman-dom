package bomberman

import (
	"errors"
	"math"
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
		case "up":
			destionation = player.YLocation + player.StepSize
			if len(g.Panel) <= destionation || g.Panel[destionation][player.Column].IsWall || g.Panel[destionation][player.Column].IsDestructible {
				return false
			}
			player.Row = destionation
			player.XLocation += player.StepSize
		case "down":
			destionation = int(math.Floor(float64(player.Row) - player.StepSize))
			if g.Players[playerIndex].Row-1 < 0 || g.Panel[g.Players[playerIndex].Row-1][g.Players[playerIndex].Column].IsWall || g.Panel[g.Players[playerIndex].Row-1][g.Players[playerIndex].Column].IsDestructible {
				return false
			}
			g.Players[playerIndex].Row--
		case "right":
			if len(g.Panel[0]) <= g.Players[playerIndex].Column+1 || g.Panel[g.Players[playerIndex].Row][g.Players[playerIndex].Column+1].IsWall || g.Panel[g.Players[playerIndex].Row][g.Players[playerIndex].Column+1].IsDestructible {
				return false
			}
			g.Players[playerIndex].Column++
		case "left":
			if g.Players[playerIndex].Column-1 < 0 || g.Panel[g.Players[playerIndex].Row][g.Players[playerIndex].Column-1].IsWall || g.Panel[g.Players[playerIndex].Row][g.Players[playerIndex].Column-1].IsDestructible {
				return false
			}
			g.Players[playerIndex].Column--
		default:
			return false
		}
	}

	return true
}
