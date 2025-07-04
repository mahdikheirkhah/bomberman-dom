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
	XLocation     int       `json:"xlocation"`
	YLocation     int       `json:"yLocation"`
	ExplosionTime time.Time `json:"explosionTime"`
	// RowRange      int       `json:"rowRange"`
	// ColRange      int       `json:"colRange"`
}

type Powerup struct {
	Type     string `json:"type"`
	Value    int    `json:"value"`
	Row      int    `json:"row"`
	Column   int    `json:"column"`
	IsHidden bool   `json:"isHidden"`
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
	return g.Players[playerIndex].NumberOfUsedBombs < g.Players[playerIndex].NumberOfBombs
}

func (g *GameBoard) CreateBomb(playerIndex int) (error, int) {
	if !g.CanCreateBomb(playerIndex) {
		return errors.New("can not create a new bomb"), -1
	}

	g.Players[playerIndex].NumberOfUsedBombs++
	var bomb Bomb
	bomb.Column = g.Players[playerIndex].Column
	bomb.Row = g.Players[playerIndex].Row
	bomb.XLocation, bomb.YLocation = g.FindGridCenterLocation(bomb.Row, bomb.Column)
	bomb.ExplosionTime = time.Now().Add(g.Players[playerIndex].BombDelay)
	g.Mu.Lock()
	g.Bombs = append(g.Bombs, bomb)
	bombIndex := len(g.Bombs) - 1
	g.Mu.Unlock()
	return nil, bombIndex
}

func (g *GameBoard) FindBombRange(bombIndex int) []Position {
	bomb := g.Bombs[bombIndex]
	var changedLocations []Position

	// Up
	for row := bomb.Row; row >= 0 && bomb.Row-row <= BombRowRange; row-- {
		cell := &g.Panel[row][bomb.Column]
		if *cell == "D" || *cell == "W" {
			if *cell == "D" {
				*cell = ""
				changedLocations = append(changedLocations, Position{Row: row, Col: bomb.Column})
			}
			break
		} else {
			*cell = "Ex"
			changedLocations = append(changedLocations, Position{Row: row, Col: bomb.Column})
		}
	}

	// Down
	for row := bomb.Row + 1; row < len(g.Panel) && row-bomb.Row <= BombRowRange; row++ {
		cell := &g.Panel[row][bomb.Column]
		if *cell == "D" || *cell == "W" {
			if *cell == "D" {
				*cell = ""
				changedLocations = append(changedLocations, Position{Row: row, Col: bomb.Column})
			}
			break
		} else {
			*cell = "Ex"
			changedLocations = append(changedLocations, Position{Row: row, Col: bomb.Column})
		}
	}

	// Left
	for col := bomb.Column - 1; col >= 0 && bomb.Column-col <= BombColRange; col-- {
		cell := &g.Panel[bomb.Row][col]
		if *cell == "D" || *cell == "W" {
			if *cell == "D" {
				*cell = ""
				changedLocations = append(changedLocations, Position{Row: bomb.Row, Col: col})
			}
			break
		} else {
			*cell = "Ex"
			changedLocations = append(changedLocations, Position{Row: bomb.Row, Col: col})
		}
	}

	// Right
	for col := bomb.Column + 1; col < len(g.Panel[0]) && col-bomb.Column <= BombColRange; col++ {
		cell := &g.Panel[bomb.Row][col]
		if *cell == "D" || *cell == "W" {
			if *cell == "D" {
				*cell = ""
				changedLocations = append(changedLocations, Position{Row: bomb.Row, Col: col})
			}
			break
		} else {
			*cell = "Ex"
			changedLocations = append(changedLocations, Position{Row: bomb.Row, Col: col})
		}
	}

	// Add the bomb's own position
	g.Panel[bomb.Row][bomb.Column] = "Ex"
	changedLocations = append(changedLocations, Position{Row: bomb.Row, Col: bomb.Column})
	g.CheckExplosion()
	return changedLocations
}

func (g *GameBoard) StartBombWatcher() {
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond) // check every 100ms
		defer ticker.Stop()

		for range ticker.C {
			g.checkBombs()
		}
	}()
}

// internal logic
func (g *GameBoard) checkBombs() {
	var remainingBombs []Bomb

	for i := 0; i < len(g.Bombs); i++ {
		bomb := g.Bombs[i]
		if time.Now().After(bomb.ExplosionTime) {
			// explode the bomb
			g.FindBombRange(i)
			// skip adding to remainingBombs
		} else {
			remainingBombs = append(remainingBombs, bomb)
		}
	}
	g.Bombs = remainingBombs
}
