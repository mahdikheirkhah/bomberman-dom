package bomberman

import (
	"errors"
	"log"
	"time"
)

const BombExplosionDuration = 1 * time.Second
const PlayerInvulnerabilityDuration = 1 * time.Second // How long player is invulnerable after respawn

// Player struct (add these fields if they're not there)

// Position represents a row and column in the game grid.
type Position struct {
	Row        int  `json:"row"`
	Col        int  `json:"col"`
	CellOnFire bool `json:"CellOnFire"`
}

// Bomb represents a bomb placed on the game board.
type Bomb struct {
	Row                 int       `json:"row"`
	Column              int       `json:"column"`
	XLocation           int       `json:"xlocation"`
	YLocation           int       `json:"yLocation"`
	ExplosionTime       time.Time `json:"explosionTime"`
	OwnPlayerIndex      int       `json:"playerIndex"`
	InitialIntersection bool      `json:"initialIntersection"`
}

// ExplodedCellInfo tracks a cell that has been exploded and its scheduled clear time.
type ExplodedCellInfo struct {
	Position  Position
	ClearTime time.Time // When this cell should revert from "Ex" to ""
}

type ExploadeCellsMsg struct {
	MsgType   string     `json:"MT"`
	Positions []Position `json:"positions"`
	BombRow   int        `json:"bombRow"`
	BombCol   int        `json:"bombCol"`
}

type PlayerExplosionMsg struct {
	Type        string `json:"type"`
	Lives       int    `json:"lives"`
	Color       string `json:"color"`
	PlayerIndex int    `json:"playerIndex"`
}
type PLayerDeath struct {
	Type   string `json:"type"`
	Player Player `json:"player"`
}

type PlantBomb struct {
	MsgType     string `json:"MT"`
	XLocation   int    `json:"XL"`
	YLocation   int    `json:"YL"`
	Row         int    `json:"R"`
	Column      int    `json:"C"`
	PlayerIndex int    `json:"PI"`
}

func (g *GameBoard) HandleBombMessage(msgMap map[string]interface{}) {
	playerIndex, ok := msgMap["fromPlayer"].(int)
	if !ok {
		log.Println("fromPlayer not found in message")
		return
	}
	g.Mu.Lock()
	defer g.Mu.Unlock()
	bombIndex, err := g.CreateBomb(playerIndex)
	if err != nil {
		log.Println("Error creating bomb:", err)
		return
	}
	var msg PlantBomb
	msg.MsgType = "BA" //Bomb Accepted
	msg.Column = g.Bombs[bombIndex].Column
	msg.Row = g.Bombs[bombIndex].Row
	msg.XLocation = g.Bombs[bombIndex].XLocation
	msg.YLocation = g.Bombs[bombIndex].YLocation
	msg.PlayerIndex = playerIndex
	g.SendMsgToChannel(msg, playerIndex)
}

func (g *GameBoard) RespawnPlayer(playerIndex int) {
	g.Players[playerIndex].Row = g.Players[playerIndex].InitialRow
	g.Players[playerIndex].Column = g.Players[playerIndex].InitialColumn
	g.Players[playerIndex].XLocation = g.Players[playerIndex].Column * g.CellSize
	g.Players[playerIndex].YLocation = g.Players[playerIndex].Row * g.CellSize
	g.Players[playerIndex].JustRespawned = true
	g.Players[playerIndex].LastDamageTime = time.Now() // Reset damage time on respawn
	time.AfterFunc(PlayerInvulnerabilityDuration, func() {
		g.Mu.Lock()
		g.Players[playerIndex].JustRespawned = false
		g.Mu.Unlock()
	})
}

// PlayerHitByExplosion checks if a player is currently within any of the given explosion positions.
// This is for immediate blast damage.
func (g *GameBoard) PlayerHitByExplosion(playerIndex int, affectedPositions []Position) bool {
	player := &g.Players[playerIndex]
	if player.IsDead || player.JustRespawned {
		return false
	}

	playerCenterRow := (player.YLocation + PlayerSize/2) / int(g.CellSize)
	playerCenterCol := (player.XLocation + PlayerSize/2) / int(g.CellSize)

	for _, pos := range affectedPositions {
		if playerCenterRow == pos.Row && playerCenterCol == pos.Col {
			return true
		}
	}
	return false
}

// DamagePlayer handles the logic for a player taking damage.
func (g *GameBoard) DamagePlayer(playerIndex int) {
	player := &g.Players[playerIndex]
	if player.IsDead || player.JustRespawned {
		return
	}

	// NEW: Check if player was recently damaged to prevent spamming
	if time.Since(player.LastDamageTime) < BombExplosionDuration/2 { // Small cooldown
		return
	}

	player.Lives--
	log.Printf("Player %d hit! Lives remaining: %d\n", playerIndex, player.Lives)
	player.LastDamageTime = time.Now() // Update last damage time

	if player.IsMoving {
		player.IsMoving = false
		if player.StopMoveChan != nil {
			close(player.StopMoveChan)
			player.StopMoveChan = nil
		}
	}

	if player.Lives <= 0 {
		g.PlayerDeath(playerIndex)
	} else {
		g.PendingRespawns = append(g.PendingRespawns, PlayerRespawn{
			PlayerIndex: playerIndex,
			RespawnTime: time.Now().Add(BombExplosionDuration),
		})

		msg := PlayerExplosionMsg{
			Type:        "PLD",
			Lives:       player.Lives,
			Color:       player.Color,
			PlayerIndex: playerIndex,
		}
		g.SendMsgToChannel(msg, -1)
	}
}

// PeriodicPlayerDamageCheck (NEW FUNCTION)
// This function will be called repeatedly by the bomb watcher.
func (g *GameBoard) PeriodicPlayerDamageCheck() {
	g.Mu.Lock()
	defer g.Mu.Unlock()
	// Loop through all players
	for i := range g.Players {
		player := &g.Players[i]
		if player.IsDead || player.JustRespawned {
			continue // Skip dead or invulnerable players
		}

		// Check if player is on an 'Ex' (exploded) cell
		// We use the player's current grid cell for this check
		playerCellRow := (player.YLocation + PlayerSize/2) / int(g.CellSize)
		playerCellCol := (player.XLocation + PlayerSize/2) / int(g.CellSize)

		if playerCellRow >= 0 && playerCellRow < NumberOfRows &&
			playerCellCol >= 0 && playerCellCol < NumberOfColumns {

			if g.Panel[playerCellRow][playerCellCol] == "Ex" {
				// Player is on an 'Ex' cell, now check if they should take damage
				// Use BombExplosionDuration as a general cooldown. This means a player
				// will only take damage from a fire cell once per full explosion duration,
				// even if they run back and forth.
				if time.Since(player.LastDamageTime) > BombExplosionDuration {
					log.Printf("Player %d walked into fire at [%d,%d]! Applying damage.", i, playerCellRow, playerCellCol)
					g.DamagePlayer(i)
				}
			}
		}
	}
}

func (g *GameBoard) PlayerDeath(playerIndex int) {
	g.NumberOfPlayers--
	g.Players[playerIndex].IsDead = true

	msg := PLayerDeath{
		Type:   "PD",
		Player: g.Players[playerIndex],
	}
	g.SendMsgToChannel(msg, -1)
	g.CheckGameEnd()
}

func (g *GameBoard) HasExploaded(row, col int) bool {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	if row < 0 || row >= len(g.Panel) || col < 0 || col >= len(g.Panel[0]) {
		return false
	}
	return g.Panel[row][col] == "Ex"
}

func (g *GameBoard) CanCreateBomb(playerIndex int) bool {

	if playerIndex < 0 || playerIndex >= len(g.Players) {
		return false
	}
	return g.Players[playerIndex].NumberOfUsedBombs < g.Players[playerIndex].NumberOfBombs
}

func (g *GameBoard) CreateBomb(playerIndex int) (int, error) {

	if playerIndex < 0 || playerIndex >= len(g.Players) {
		return -1, errors.New("invalid player index")
	}

	if !g.CanCreateBomb(playerIndex) {
		return -1, errors.New("can not create a new bomb: player has reached bomb limit")
	}

	g.Players[playerIndex].NumberOfUsedBombs++

	var bomb Bomb
	bomb.Column = g.Players[playerIndex].Column
	bomb.Row = g.Players[playerIndex].Row
	bomb.XLocation = bomb.Column * g.CellSize
	bomb.YLocation = bomb.Row * g.CellSize
	bomb.ExplosionTime = time.Now().Add(g.Players[playerIndex].BombDelay)
	bomb.OwnPlayerIndex = playerIndex
	bomb.InitialIntersection = true

	g.Bombs = append(g.Bombs, bomb)
	bombIndex := len(g.Bombs) - 1
	return bombIndex, nil
}

func (g *GameBoard) CalculateBombRange(bombRow, bombCol, bombRange int) []Position {
	var affectedPositions []Position

	affectedPositions = append(affectedPositions, Position{Row: bombRow, Col: bombCol})

	for row := bombRow - 1; row >= 0 && bombRow-row <= bombRange; row-- {
		if g.Panel[row][bombCol] == "W" {
			break
		}
		affectedPositions = append(affectedPositions, Position{Row: row, Col: bombCol})
		if g.Panel[row][bombCol] == "D" {
			g.CreatePowerupWithChance(row, bombCol) // Uncomment if you have this function
			break
		}
	}

	for row := bombRow + 1; row < len(g.Panel) && row-bombRow <= bombRange; row++ {
		if g.Panel[row][bombCol] == "W" {
			break
		}
		affectedPositions = append(affectedPositions, Position{Row: row, Col: bombCol})
		if g.Panel[row][bombCol] == "D" {
			g.CreatePowerupWithChance(row, bombCol) // Uncomment if you have this function
			break
		}
	}

	for col := bombCol - 1; col >= 0 && bombCol-col <= bombRange; col-- {
		if g.Panel[bombRow][col] == "W" {
			break
		}
		affectedPositions = append(affectedPositions, Position{Row: bombRow, Col: col})
		if g.Panel[bombRow][col] == "D" {
			g.CreatePowerupWithChance(bombRow, col) // Uncomment if you have this function
			break
		}
	}

	for col := bombCol + 1; col < len(g.Panel[0]) && col-bombCol <= bombRange; col++ {
		if g.Panel[bombRow][col] == "W" {
			break
		}
		affectedPositions = append(affectedPositions, Position{Row: bombRow, Col: col})
		if g.Panel[bombRow][col] == "D" {
			g.CreatePowerupWithChance(bombRow, col) // Uncomment if you have this function
			break
		}
	}

	return affectedPositions
}

func (g *GameBoard) ApplyExplosion(bomb Bomb) {
	var player *Player
	for i := range g.Players {
		if i == bomb.OwnPlayerIndex {
			player = &g.Players[i]
			break
		}
	}
	if player == nil {
		log.Printf("Player with index %d not found for bomb explosion.", bomb.OwnPlayerIndex)
		return
	}

	affectedPositions := g.CalculateBombRange(bomb.Row, bomb.Column, player.BombRange)

	for i := range g.Bombs {
		if g.Bombs[i].Row == bomb.Row && g.Bombs[i].Column == bomb.Column {
			continue
		}

		for _, pos := range affectedPositions {
			if g.Bombs[i].Row == pos.Row && g.Bombs[i].Column == pos.Col {
				g.Bombs[i].ExplosionTime = time.Now()
				break
			}
		}
	}

	var msg ExploadeCellsMsg
	msg.MsgType = "EXC"
	msg.BombRow = bomb.Row
	msg.BombCol = bomb.Column
	for _, pos := range affectedPositions {
		if pos.Row < 0 || pos.Row >= len(g.Panel) || pos.Col < 0 || pos.Col >= len(g.Panel[0]) {
			continue
		}

		cell := &g.Panel[pos.Row][pos.Col]
		if *cell != "W" {
			*cell = "Ex"
			g.ExplodedCells = append(g.ExplodedCells, ExplodedCellInfo{Position: pos, ClearTime: time.Now().Add(BombExplosionDuration)})
			msg.Positions = append(msg.Positions, Position{Row: pos.Row, Col: pos.Col, CellOnFire: true})
		}
	}
	g.SendMsgToChannel(msg, -1)

	// IMMEDIATE DAMAGE: Check and damage players caught in THIS SPECIFIC explosion.
	for i := range g.Players {
		if g.PlayerHitByExplosion(i, affectedPositions) {
			g.DamagePlayer(i)
		}
	}

	player.NumberOfUsedBombs--
}

func (g *GameBoard) ClearExpiredExplosions() {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	var remainingExplodedCells []ExplodedCellInfo
	now := time.Now()
	var msg ExploadeCellsMsg
	for _, info := range g.ExplodedCells {
		if now.After(info.ClearTime) {
			if info.Position.Row >= 0 && info.Position.Row < len(g.Panel) &&
				info.Position.Col >= 0 && info.Position.Col < len(g.Panel[0]) {
				if g.Panel[info.Position.Row][info.Position.Col] == "Ex" {
					powerupIndex := g.FindPowerupAt(info.Position.Row, info.Position.Col)
					if powerupIndex != -1 {
						if g.Powerups[powerupIndex].IsHidden {
							g.ShowPowerup(powerupIndex)
						} else {
							g.RemovePowerup(powerupIndex)
						}
					}
					// g.FindPowerupAt, g.RemovePowerup, g.ShowPowerup calls here if you have them
					g.Panel[info.Position.Row][info.Position.Col] = ""
					msg.MsgType = "OF"
					msg.Positions = append(msg.Positions, Position{Row: info.Position.Row, Col: info.Position.Col})
				}
			}
		} else {
			remainingExplodedCells = append(remainingExplodedCells, info)
		}
	}
	if msg.MsgType == "OF" {
		g.SendMsgToChannel(msg, -1)
	}
	g.ExplodedCells = remainingExplodedCells
}

func (g *GameBoard) ProcessRespawns() {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	var remainingRespawns []PlayerRespawn
	now := time.Now()

	for _, respawn := range g.PendingRespawns {
		if now.After(respawn.RespawnTime) {
			g.RespawnPlayer(respawn.PlayerIndex)
			player := g.Players[respawn.PlayerIndex]
			msg := struct {
				Type        string `json:"type"`
				PlayerIndex int    `json:"playerIndex"`
				XLocation   int    `json:"xlocation"`
				YLocation   int    `json:"yLocation"`
			}{
				Type:        "PR",
				PlayerIndex: respawn.PlayerIndex,
				XLocation:   player.XLocation,
				YLocation:   player.YLocation,
			}
			g.SendMsgToChannel(msg, -1)
		} else {
			remainingRespawns = append(remainingRespawns, respawn)
		}
	}
	g.PendingRespawns = remainingRespawns
}

func (g *GameBoard) StartBombWatcher() {
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond) // Check every 100ms
		defer ticker.Stop()

		for range ticker.C {
			//g.Mu.Lock() // Lock for the entire loop iteration
			g.checkBombs()
			g.ClearExpiredExplosions()
			g.ProcessRespawns()
			g.PeriodicPlayerDamageCheck() // NEW: Call the periodic damage check here
			//g.Mu.Unlock()
		}
	}()
}

func (g *GameBoard) checkBombs() {
	g.Mu.Lock()
	defer g.Mu.Unlock()
	// This function already acquires its own lock, so no need for g.Mu.Lock() around it in StartBombWatcher
	// However, if you're calling it from StartBombWatcher within a g.Mu.Lock() block,
	// it would cause a deadlock. Let's adjust StartBombWatcher to lock around the *whole* tick.
	// (See StartBombWatcher definition above)

	var remainingBombs []Bomb
	now := time.Now()

	for _, bomb := range g.Bombs {
		if now.After(bomb.ExplosionTime) {
			g.ApplyExplosion(bomb)
		} else {
			remainingBombs = append(remainingBombs, bomb)
		}
	}
	g.Bombs = remainingBombs
}
