package bomberman

import (
	"log"
	"math/rand"
)



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
	// 30% chance to create a powerup
	if rand.Float64() > 0.3 {
		return
	}

	// Check if a powerup already exists at this location
	if g.FindPowerupAt(row, column) != -1 {
		log.Println("Powerup already exists at this location")
		return
	}

	// Calculate weights for each powerup type
	// Powerups chosen less frequently will have higher weights
	weights := make(map[string]float64)
	totalWeight := 0.0

	// Calculate inverse frequency as weight. Add 1 to denominator to avoid division by zero
	// and to ensure even types with 0 count have a weight.
	for _, pType := range PowerupTypes {
		// The lower the count, the higher the weight
		weight := 1.0 / float64(g.powerupChosen[pType]+1)
		weights[pType] = weight
		totalWeight += weight
	}

	// Generate a random number within the total weight
	r := rand.Float64() * totalWeight

	// Select the powerup type based on weighted chance
	var chosenType string
	currentWeightSum := 0.0
	for _, pType := range PowerupTypes {
		currentWeightSum += weights[pType]
		if r <= currentWeightSum {
			chosenType = pType
			break
		}
	}

	// If for some reason no type was chosen (shouldn't happen with correct weights),
	// fall back to a simple random selection.
	if chosenType == "" {
		chosenType = PowerupTypes[rand.Intn(len(PowerupTypes))]
	}

	// Create the powerup instance
	var powerup Powerup
	powerup.Row = row
	powerup.Column = column
	powerup.IsHidden = true
	powerup.Type = chosenType

	// Increment the count for the chosen powerup type
	g.powerupChosen[powerup.Type]++

	// Assign value based on type
	switch powerup.Type {
	case "ExtraBomb":
		powerup.Value = 1
	case "BombRange":
		powerup.Value = 1
	case "ExtraLife":
		powerup.Value = 1
	case "SpeedBoost":
		powerup.Value = 2
	}

	log.Println("Powerup added:", powerup)
	g.Powerups = append(g.Powerups, powerup)
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
	g.RemovePowerup(PowerupIndex)
	player := &g.Players[playerIndex]
	switch powerup.Type {
	case "ExtraBomb":
		if player.NumberOfBombs >= MaxBombsPowerup {
			return
		}
		player.NumberOfBombs += powerup.Value
	case "BombRange":
		if player.BombRange >= MaxBombRangePowerup {
			return
		}
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
		if player.StepSize >= MaxSpeedPowerup {
			return
		}
		player.StepSize += powerup.Value
	}
}

func (g *GameBoard) FindPowerupAt(row, column int) int {
	for i, powerup := range g.Powerups {
		if powerup.Row == row && powerup.Column == column {
			return i
		}
	}
	return -1
}
