package bomberman

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
