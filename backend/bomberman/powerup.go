package bomberman

import "math/rand"

// Powerup represents an item that can be collected by players.
type Powerup struct {
	Type     string `json:"type"`
	Value    int    `json:"value"`
	Row      int    `json:"row"`
	Column   int    `json:"column"`
	IsHidden bool   `json:"isHidden"`
}

var PowerupTypes = []string{"ExtraBomb", "BombRange", "ExtraLife", "SpeedBoost"}

func (g *GameBoard) ShowPowerup(PowerUpIndex int) {
	if PowerUpIndex < 0 || PowerUpIndex >= len(g.Powerups) {
		return
	}
	powerup := g.Powerups[PowerUpIndex]
	if !powerup.IsHidden {
		return // Already shown
	}
	powerup.IsHidden = false
	g.Powerups[PowerUpIndex] = powerup
	g.SendMsgToChannel(struct {
		Type    string  `json:"type"`
		Powerup Powerup `json:"powerup"`
	}{
		Type:    "AddPowerup",
		Powerup: powerup,
	}, -1)
}

func (g *GameBoard) CreatePowerupWithChance(row, column int) {
	if rand.Float64() > 0.4 { // 40% chance to create a powerup
		return
	}
	if g.FindPowerupAt(row, column) != -1 {
		return // Powerup already exists at this location
	}
	var Powerup Powerup
	Powerup.Row = row
	Powerup.Column = column
	Powerup.IsHidden = true
	Powerup.Type = PowerupTypes[rand.Intn(len(PowerupTypes))]
	switch Powerup.Type {
	case "ExtraBomb":
		Powerup.Value = 1
	case "BombRange":
		Powerup.Value = 1
	case "ExtraLife":
		Powerup.Value = 1
	case "SpeedBoost":
		Powerup.Value = 2
	}
	g.Powerups = append(g.Powerups, Powerup)
}

func (g *GameBoard) RemovePowerup(PowerupIndex int) {
	if PowerupIndex < 0 || PowerupIndex >= len(g.Powerups) {
		return
	}
	powerup := g.Powerups[PowerupIndex]
	g.Powerups = append(g.Powerups[:PowerupIndex], g.Powerups[PowerupIndex+1:]...)
	g.SendMsgToChannel(struct {
		Type   string `json:"type"`
		Row    int    `json:"row"`
		Column int    `json:"column"`
	}{
		Type:   "RemovePowerup",
		Row:    powerup.Row,
		Column: powerup.Column,
	}, -1)
}

func (g *GameBoard) EatPowerup(playerIndex, PowerupIndex int) {
	if PowerupIndex < 0 || PowerupIndex >= len(g.Powerups) {
		return
	}
	powerup := g.Powerups[PowerupIndex]
	if powerup.IsHidden {
		return
	}
	g.Powerups = append(g.Powerups[:PowerupIndex], g.Powerups[PowerupIndex+1:]...)
	player := &g.Players[playerIndex]
	switch powerup.Type {
	case "ExtraBomb":
		player.NumberOfBombs += powerup.Value
		g.SendMsgToChannel(struct {
			Type              string `json:"type"`
			Player            int    `json:"player"`
			NumberOfBombs     int    `json:"numberOfBombs"`
			NumberOfUsedBombs int    `json:"numberOfUsedBombs"`
		}{
			Type:              "EatBombPowerup",
			Player:            playerIndex,
			NumberOfBombs:     player.NumberOfBombs,
			NumberOfUsedBombs: player.NumberOfUsedBombs,
		}, -1)
	case "BombRange":
		player.BombRange += powerup.Value
	case "ExtraLife":
		player.Lives += powerup.Value
		g.SendMsgToChannel(struct {
			Type          string `json:"type"`
			Player        int    `json:"player"`
			NumberOfLives int    `json:"numberOfLives"`
		}{
			Type:          "EatLifePowerup",
			Player:        playerIndex,
			NumberOfLives: player.Lives,
		}, -1)
	case "SpeedBoost":
		player.StepSize += powerup.Value
	}
}

func (g *GameBoard) FindPowerupAt(row, column int) int {
	for i, powerup := range g.Powerups {
		if powerup.Row == row && powerup.Column == column && !powerup.IsHidden {
			return i
		}
	}
	return -1
}
