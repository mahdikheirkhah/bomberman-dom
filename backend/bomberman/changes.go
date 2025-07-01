package bomberman

import (
	"errors"
	"time"
)

const BombRowRange = 2
const BombColRange = 2

type Bomb struct {
	Row           int       `json:"row"`
	Column        int       `json:"column"`
	ExplosionTime time.Time `json:"explosionTime"`
	// RowRange      int       `json:"rowRange"`
	// ColRange      int       `json:"colRange"`
}

func (g *GameBoard) CheckExplosion() {
	for i, player := range g.Players {
		if g.HasExploaded(player.Row, player.Column) {
			g.Players[i].Lives--
		}

		if player.Lives == 0 {
			g.NumberOfPlayers--
			g.Players[i].IsDead = true
		}
	}

}

func (g *GameBoard) CanCreateBomb(playerIndex int) bool {
	if g.Players[playerIndex].NumberOfUsedBombs >= g.Players[playerIndex].NumberOfBombs {
		return false
	}
	return true
}

func (g *GameBoard) CreateBomb(playerIndex int) error {
	if !g.CanCreateBomb(playerIndex) {
		return errors.New("Can not create a new bomb")
	}

	g.Players[playerIndex].NumberOfUsedBombs++
	var bomb Bomb
	bomb.Column = g.Players[playerIndex].Column
	bomb.Row = g.Players[playerIndex].Row
	bomb.ExplosionTime = time.Now().Add(g.Players[playerIndex].BombDelay)
	g.Bombs = append(g.Bombs, bomb)
	return nil
}

func (g *GameBoard) FindBombRange(bombIndex int) (int, int, int, int) {
	bomb := g.Bombs[bombIndex]
	var startRow, endRow, startCol, endCol int

	for row := bomb.Row; row >= 0 && bomb.Row-row <= BombRowRange; row-- {
		if !(g.Panel[row][bomb.Column].IsDestructible || g.Panel[row][bomb.Column].IsWall) {
			startRow = row
		} else {
			break
		}
	}
	for row := bomb.Row; row < len(g.Panel) && row-bomb.Row <= BombRowRange; row++ {
		if !(g.Panel[row][bomb.Column].IsDestructible || g.Panel[row][bomb.Column].IsWall) {
			endRow = row
		} else {
			break
		}

	}

	for col := bomb.Column; col >= 0 && bomb.Column-col <= BombColRange; col-- {
		if !(g.Panel[bomb.Row][col].IsDestructible || g.Panel[bomb.Row][col].IsWall) {
			startCol = col
		} else {
			break
		}
	}
	for col := bomb.Column; col < len(g.Panel[0]) && col-bomb.Column <= BombRowRange; col++ {
		if !(g.Panel[bomb.Row][col].IsDestructible || g.Panel[bomb.Row][col].IsWall) {
			startCol = col
		} else {
			break
		}

	}
	return startRow, endRow, startCol, endCol
}
