package bomberman

import (
	"errors"
	"time"
)

const BombRowRange = 2
const BombColRange = 2

type Position struct {
	Row int
	Col int
}
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

func (g *GameBoard) FindBombRange(bombIndex int) []Position {
	bomb := g.Bombs[bombIndex]
	var changedLocations []Position

	// Up
	for row := bomb.Row; row >= 0 && bomb.Row-row <= BombRowRange; row-- {
		cell := &g.Panel[row][bomb.Column]
		if cell.IsDestructible || cell.IsWall {
			if cell.IsDestructible {
				cell.IsDestructible = false
				changedLocations = append(changedLocations, Position{Row: row, Col: bomb.Column})
			}
			break
		} else {
			cell.IsExploaded = true
			changedLocations = append(changedLocations, Position{Row: row, Col: bomb.Column})
		}
	}

	// Down
	for row := bomb.Row + 1; row < len(g.Panel) && row-bomb.Row <= BombRowRange; row++ {
		cell := &g.Panel[row][bomb.Column]
		if cell.IsDestructible || cell.IsWall {
			if cell.IsDestructible {
				cell.IsDestructible = false
				changedLocations = append(changedLocations, Position{Row: row, Col: bomb.Column})
			}
			break
		} else {
			cell.IsExploaded = true
			changedLocations = append(changedLocations, Position{Row: row, Col: bomb.Column})
		}
	}

	// Left
	for col := bomb.Column - 1; col >= 0 && bomb.Column-col <= BombColRange; col-- {
		cell := &g.Panel[bomb.Row][col]
		if cell.IsDestructible || cell.IsWall {
			if cell.IsDestructible {
				cell.IsDestructible = false
				changedLocations = append(changedLocations, Position{Row: bomb.Row, Col: col})
			}
			break
		} else {
			cell.IsExploaded = true
			changedLocations = append(changedLocations, Position{Row: bomb.Row, Col: col})
		}
	}

	// Right
	for col := bomb.Column + 1; col < len(g.Panel[0]) && col-bomb.Column <= BombColRange; col++ {
		cell := &g.Panel[bomb.Row][col]
		if cell.IsDestructible || cell.IsWall {
			if cell.IsDestructible {
				cell.IsDestructible = false
				changedLocations = append(changedLocations, Position{Row: bomb.Row, Col: col})
			}
			break
		} else {
			cell.IsExploaded = true
			changedLocations = append(changedLocations, Position{Row: bomb.Row, Col: col})
		}
	}

	// Add the bomb's own position
	g.Panel[bomb.Row][bomb.Column].IsExploaded = true
	changedLocations = append(changedLocations, Position{Row: bomb.Row, Col: bomb.Column})

	return changedLocations
}

func (g *GameBoard) BombExplosion() {

	for true {
		for i, bomb := range g.Bombs {
			if time.Now().After(bomb.ExplosionTime) {
				g.FindBombRange(i)
				if i == len(g.Bombs)-1 {
					g.Bombs = g.Bombs[:i]
				} else {
					g.Bombs = append(g.Bombs[:i], g.Bombs[i+1:]...)
				}
				break
			}
		}
	}

}
