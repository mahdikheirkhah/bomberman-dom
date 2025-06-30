package bomberman

type Player struct {
	Name      string `json:"name"`
	Lives     int    `json:"lives"`
	Score     int    `json:"score"`
	Color     string `json:"color"`
	XLocation int    `json:"xLocation"`
	YLocation int    `json:"yLocation"`
}
