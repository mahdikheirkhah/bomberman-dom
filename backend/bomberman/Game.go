package bomberman

type GameBorad struct {
}

type GameCell struct {
	IsOccupied  bool `json:"isOccupied"`
	IsWall      bool `json:"isWall"`
	IsExploaded bool `json:"isExploaded"`
}
