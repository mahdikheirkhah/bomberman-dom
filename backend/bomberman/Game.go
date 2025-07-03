package bomberman

const NumberOfRows = 20
const NumberOfColumns = 20
const MaxNumberOfPlayers = 4
const MinNumberOfPlayers = 2
const GridSize = 20.00

var Colors = []string{"G", "Y", "R", "B"}

type GameBoard struct {
	Players         []Player                                `json:"players"`
	Bombs           []Bomb                                  `json:"bombs"`
	NumberOfPlayers int                                     `json:"numberOfPlayers"`
	Panel           [NumberOfRows][NumberOfColumns]GameCell `json:"panel"`
	GridSize        float64                                 `json:"gridSize"`
}

type GameCell struct {
	IsOccupied     bool `json:"isOccupied"`
	IsWall         bool `json:"isWall"`
	IsDestructible bool `json:"isDestructible"`
	IsExploaded    bool `json:"isExploaded"`
}

func (g *GameBoard) CanCreateNewPlayer() bool {
	if 0 < g.NumberOfPlayers+1 && g.NumberOfPlayers+1 <= MaxNumberOfPlayers {
		return true
	}
	return false
}

func (g *GameBoard) FindColor() string {
	return Colors[g.NumberOfPlayers+1]
}

func (g *GameBoard) FindStartRowLocation() int {
	if g.NumberOfPlayers+1 == 1 || g.NumberOfPlayers+1 == 2 {
		return 0
	}
	return NumberOfRows - 1
}

func (g *GameBoard) FindStartColLocation() int {
	if g.NumberOfPlayers+1 == 1 || g.NumberOfPlayers+1 == 3 {
		return 0
	}
	return NumberOfColumns - 1
}

func (g *GameBoard) HasExploaded(row, col int) bool {
	return g.Panel[row][col].IsExploaded
}

func (g *GameBoard) FindInnerCell(axis byte, direction byte, location float64, playerIndex int) int {
	col := g.Players[playerIndex].Column
	row := g.Players[playerIndex].Row
	if axis == 'x' {
		if direction == 'r' {
			if location > g.FindGridBorderLocation('r', playerIndex) {
				return col + 1
			}
		} else if direction == 'l' {
			if location < g.FindGridBorderLocation('l', playerIndex) {
				return col - 1
			}
		}
		return col
	} else if axis == 'y' {
		if direction == 'u' {
			if location > g.FindGridBorderLocation('u', playerIndex) {
				return row + 1
			}
		} else if direction == 'd' {
			if location < g.FindGridBorderLocation('d', playerIndex) {
				return row - 1
			}
		}
		return row
	}
	return 0
}

func (g *GameBoard) FindGridBorderLocation(borderName byte, playerIndex int) float64 {
	row := g.Players[playerIndex].Row
	col := g.Players[playerIndex].Column
	switch borderName {
	case 'u':
		return (float64(row) * g.GridSize) + g.GridSize
	case 'd':
		return float64(row) * g.GridSize
	case 'r':
		return (float64(col) * g.GridSize) + g.GridSize
	case 'l':
		return (float64(col) * g.GridSize)
	}
	return -1
}

func (g *GameBoard) FindGridCenterLocation(row, col int) (float64, float64) {
	return (float64(col) * g.GridSize) + (g.GridSize / 2), (float64(row) * g.GridSize) + (g.GridSize / 2)
}
