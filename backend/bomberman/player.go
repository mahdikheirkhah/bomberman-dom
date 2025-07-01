package bomberman

import "errors"

type Player struct {
	Name   string `json:"name"`
	Lives  int    `json:"lives"`
	Score  int    `json:"score"`
	Color  string `json:"color"`
	Row    int    `json:"row"`
	Column int    `json:"column"`
	IsDead bool   `json:"isDead"`
}

func (g *GameBoard) CreatePlayer(name string) error {
	var player Player
	if !g.CanCreateNewPlayer() {
		return errors.New("Max number of players of has been reached!")
	}

	player.Name = name
	player.Lives = 3
	player.Score = 0
	player.Color = g.FindColor()
	player.Row = g.FindStartRowLocation()
	player.Column = g.FindStartColLocation()
	g.Players = append(g.Players, player)
	g.NumberOfPlayers++

	return nil
}

func (g *GameBoard) MovePlayer(playerIndex int, direction string) bool {
	switch direction {
	case "row-forward":
		if len(g.Panel) <= g.Players[playerIndex].Row+1 || g.Panel[g.Players[playerIndex].Row+1][g.Players[playerIndex].Column].IsWall {
			return false
		}
		g.Players[playerIndex].Row++
	case "row-backward":
		if g.Players[playerIndex].Row-1 < 0 || g.Panel[g.Players[playerIndex].Row-1][g.Players[playerIndex].Column].IsWall {
			return false
		}
		g.Players[playerIndex].Row--
	case "col-upward":
		if len(g.Panel[0]) <= g.Players[playerIndex].Column+1 || g.Panel[g.Players[playerIndex].Row][g.Players[playerIndex].Column+1].IsWall {
			return false
		}
		g.Players[playerIndex].Column++
	case "col-downward":
		if g.Players[playerIndex].Column-1 < 0 || g.Panel[g.Players[playerIndex].Row][g.Players[playerIndex].Column-1].IsWall {
			return false
		}
		g.Players[playerIndex].Column--
	default:
		return false
	}
	return true
}
