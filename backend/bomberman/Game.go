package bomberman

const NumberOfRows = 20
const NumberOfColumns = 20
const MaxNumberOfPlayers = 4
const MinNumberOfPlayers = 2

var Colors = []string{"G", "Y", "R", "B"}

type GameBoard struct {
	Players         []Player                                `json:"players"`
	Bombs           []Bomb                                  `json:"bombs"`
	NumberOfPlayers int                                     `json:"numberOfPlayers"`
	Panel           [NumberOfRows][NumberOfColumns]GameCell `json:"panel"`
	GridSize        int                                     `json:"gridSize"`
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

func (g *GameBoard) FindInnerCell(axis byte, location float64, playerIndex int) int {
	if axis == 'x' {

	} else if axis == 'y' {

	}
	return -1
}

func (g *GameBoard) FindGridBorderLocation(borderName byte, playerIndex int) int {
	row := g.Players[playerIndex].Row
	col := g.Players[playerIndex].Column
	switch borderName {
	case 'u':
		return (col * g.GridSize) + g.GridSize
	case 'd':
		return col * g.GridSize
	case 'l':
		return (row * g.GridSize) + g.GridSize
	case 'r':
		return (row * g.GridSize)
	}
	return -100
}

func (g *GameBoard) FindGridCenterLocation(axis byte, cell int) int {

}
