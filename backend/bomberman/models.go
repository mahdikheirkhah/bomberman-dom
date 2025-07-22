package bomberman

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"net/http"
)

// Game.go
const NumberOfRows = 11
const NumberOfColumns = 13
const MaxNumberOfPlayers = 4
const MinNumberOfPlayers = 2
const CellSize = 50

var lobbyCountdownTimer = 20 // 20 for production
var startCountdownTimer = 10 // 10 for production

var Colors = []string{"G", "Y", "R", "B"}

type PlayerRespawn struct {
	PlayerIndex int
	RespawnTime time.Time
}

type GameBoard struct {
	Players            []Player                              `json:"players"`
	Bombs              []Bomb                                `json:"bombs"`
	PendingRespawns    []PlayerRespawn                       `json:"-"`
	NumberOfPlayers    int                                   `json:"numberOfPlayers"`
	Panel              [NumberOfRows][NumberOfColumns]string `json:"panel"` // Ex -> Explode , W -> Wall, D -> Destructible, ""(empty) -> empty cell, B -> Bomb
	CellSize           int                                   `json:"cellSize"`
	Powerups           []Powerup                             `json:"powerups"`
	IsStarted          bool
	GameState          string // lobby, gameCountdown, gameStarted
	StopCountdown      bool
	ExplodedCells      []ExplodedCellInfo `json:"explodedCells"`
	PlayersConnections map[int]*websocket.Conn
	powerupChosen      map[string]int
	BroadcastChannel   chan interface{}

	Mu sync.Mutex
}

type GameCell struct {
	IsOccupied     bool `json:"isOccupied"`
	IsWall         bool `json:"isWall"`
	IsDestructible bool `json:"isDestructible"`
	IsExploaded    bool `json:"isExploaded"`
}

// bomb.go
const BombExplosionDuration = 1 * time.Second
const PlayerInvulnerabilityDuration = 1 * time.Second // How long player is invulnerable after respawn

type Position struct {
	Row        int  `json:"row"`
	Col        int  `json:"col"`
	CellOnFire bool `json:"CellOnFire"`
}

type Bomb struct {
	Row                 int       `json:"row"`
	Column              int       `json:"column"`
	XLocation           int       `json:"xlocation"`
	YLocation           int       `json:"yLocation"`
	ExplosionTime       time.Time `json:"explosionTime"`
	OwnPlayerIndex      int       `json:"playerIndex"`
	InitialIntersection bool      `json:"initialIntersection"`
}

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

// broadcast.go
var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections
	},
}

type JoinRequest struct {
	Name string `json:"name"`
}

type StateMsg struct {
	Type  string `json:"type"`
	State string `json:"state"`
}

var LobbyMsg bool

// chatMsg.go
type Chat struct {
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	Content     string    `json:"content"`
	Date        time.Time `json:"date"`
	Filter      bool      `json:"filter"`
	SenderIndex int       `json:"senderIndex"`
	Color       string    `json:"color"`
}

// move.go
const movementTolerance = 20

type MovePlayerMsg struct {
	MsgType     string `json:"MT"`
	XLocation   int    `json:"XL"`
	YLocation   int    `json:"YL"`
	Direction   string `json:"D"`
	PlayerIndex int    `json:"PI"`
}

// player.go
const StepSize = 5
const BombDelay = 3
const BombRange = 2
const PlayerSize = 48

type Player struct {
	Index             int           `json:"index"`
	Name              string        `json:"name"`
	Lives             int           `json:"lives"`
	Score             int           `json:"score"`
	Color             string        `json:"color"`
	Row               int           `json:"row"`
	Column            int           `json:"column"`
	InitialRow        int           `json:"initialRow"`
	InitialColumn     int           `json:"initialColumn"`
	XLocation         int           `json:"xlocation"`
	YLocation         int           `json:"yLocation"`
	IsDead            bool          `json:"isDead"`
	IsHurt            bool          `json:"isHurt"`
	NumberOfBombs     int           `json:"numberOfBombs"`
	NumberOfUsedBombs int           `json:"numberOfUsedBombs"`
	BombDelay         time.Duration `json:"-"`
	BombRange         int           `json:"bombRange"`
	StepSize          int           `json:"stepSize"`
	DirectionFace     string        `json:"DirectionFace"`
	IsMoving          bool          `json:"isMoving"`
	JustRespawned     bool          `json:"justRespawned"`
	LastDamageTime    time.Time     `json:"lastDamageTime"`
	UUID              string        `json:"uuid"`
	StopMoveChan      chan struct{} `json:"-"` // Channel to signal the player's movement goroutine to stop
}

// powerup.go
const MaxBombsPowerup = 5
const MaxBombRangePowerup = 5
const MaxSpeedPowerup = 20

var PowerupTypes = []string{"ExtraBomb", "BombRange", "ExtraLife", "SpeedBoost"}

type Powerup struct {
	Type     string `json:"type"`
	Value    int    `json:"value"`
	Row      int    `json:"row"`
	Column   int    `json:"column"`
	IsHidden bool   `json:"isHidden"`
}
