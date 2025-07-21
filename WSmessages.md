# Bomberman WebSocket API Documentation

This document outlines the WebSocket messages used for communication between the game client (frontend) and the game server (backend).

## General Concepts

- **Client-to-Server (C->S):** Messages sent from the player's browser to the server.
- **Server-to-Client (S->C):** Messages sent from the server to one or more players' browsers.
- All messages are JSON objects.
- Server-to-client messages are identified by either a `type` field or an `MT` (MessageType) field.

---

## Client-to-Server (C->S) Messages

These messages are sent from the client to the server to report player actions. They use the `msgType` field.

### `MS` (Move Start)
- **Description:** Sent when a player presses a movement key to start moving.
- **Payload:**
  ```json
  {
    "msgType": "MS",
    "d": "u"
  }
  ```
- **Fields:**
  - `msgType` (string): `"MS"`
  - `d` (string): Direction of movement: `"u"`, `"d"`, `"l"`, or `"r"`.

### `ME` (Move End)
- **Description:** Sent when the player releases the last movement key to stop moving.
- **Payload:**
  ```json
  {
    "msgType": "ME"
  }
  ```

### `b` (Place Bomb)
- **Description:** Sent when the player presses the space key to place a bomb.
- **Payload:**
  ```json
  {
    "msgType": "b"
  }
  ```

### `c` (Chat Message)
- **Description:** Sent when a player submits a chat message.
- **Payload:**
  ```json
  {
    "msgType": "c",
    "content": "Hello!"
  }
  ```
- **Fields:**
  - `content` (string): The text of the message.

---

## Server-to-Client (S->C) Messages

These messages are sent from the server to the client(s) to update the game state.

### Messages with `type` field

#### `PlayerAccepted`
- **Description:** Confirms to a client that they have successfully joined the game.
- **Payload:** `{"type":"PlayerAccepted","index":0}`
- **Fields:**
  - `index` (number): The player's assigned index.

#### `player_list`
- **Description:** Provides the current list of players in the lobby.
- **Payload:** `{"type":"player_list","players":[...]}`

#### `GameState`
- **Description:** Informs the client of a major change in the game's state.
- **Payload:** `{"type":"GameState","state":"LobbyCountdown"}`
- **Possible States:** `LobbyCountdown`, `GameCountdown`, `GameStarted`, `GameOver`, `StopCountdown`.

#### `lobbyCountdown` / `gameCountdown`
- **Description:** Provides the remaining seconds in a countdown.
- **Payload:** `{"type":"lobbyCountdown","seconds":5}`

#### `gameStart`
- **Description:** Sent when the game begins, containing the initial board and player data.
- **Payload:** `{"type":"gameStart","players":[...],"panel":[[...]]}`

#### `CM` (Chat Message)
- **Description:** Broadcasts a chat message to all players.
- **Payload:** `{"type":"CM", "name":"player1", "content":"Hi!", "date":"...", "filter":false, "senderIndex":0, "color":"G"}`

#### `PlayerDisconnected`
- **Description:** Sent when a player disconnects from the game.
- **Payload:** `{"type":"PlayerDisconnected","index":1}`

#### `PD` (Player Death)
- **Description:** Sent when a player has lost all their lives.
- **Payload:** `{"type":"PD", "player":{...}}`

#### `PLD` (Player Lives Damaged)
- **Description:** Sent when a player is damaged by an explosion.
- **Payload:** `{"type":"PLD", "lives":2, "color":"G", "playerIndex":0}`

#### `PR` (Player Respawn)
- **Description:** Sent when a player respawns after being damaged.
- **Payload:** `{"type":"PR", "playerIndex":0, "xlocation":50, "yLocation":50}`

#### `AddPowerup`
- **Description:** Sent when a power-up is revealed on the map.
- **Payload:** `{"type":"AddPowerup", "powerup":{...}}`

#### `RemovePowerup`
- **Description:** Sent when a power-up is collected or removed.
- **Payload:** `{"type":"RemovePowerup", "row":3, "column":5}`

#### `EatLifePowerup`
- **Description:** Sent when a player collects a life power-up.
- **Payload:** `{"type":"EatLifePowerup", "player":0, "numberOfLives":4}`

### Messages with `MT` (MessageType) field

#### `M` (Player Move)
- **Description:** Broadcasts a player's new position and direction.
- **Payload:** `{"MT":"M","PI":0,"XL":150,"YL":50,"D":"r"}`
- **Fields:**
  - `PI` (number): Player Index.
  - `XL` (number): X Location (pixels).
  - `YL` (number): Y Location (pixels).
  - `D` (string): Direction.

#### `BA` (Bomb Accepted)
- **Description:** Confirms a bomb has been placed and provides its location.
- **Payload:** `{"MT":"BA", "XL":100, "YL":50, "R":1, "C":2, "PI":0}`
- **Fields:**
  - `XL`, `YL` (number): Pixel coordinates of the bomb.
  - `R`, `C` (number): Row and column index of the bomb.
  - `PI` (number): Player Index of the bomb owner.

#### `EXC` (Explosion)
- **Description:** Sent when a bomb explodes, indicating the affected cells.
- **Payload:** `{"MT":"EXC", "positions":[{"row":1, "col":2, "CellOnFire":true}, ...], "bombRow":1, "bombCol":2}`
- **Fields:**
  - `positions` (array): A list of cells affected by the explosion.
  - `bombRow`, `bombCol` (number): The location of the bomb that exploded.

#### `OF` (Off Fire)
- **Description:** Sent to clear the fire from the grid after an explosion.
- **Payload:** `{"MT":"OF", "positions":[{"row":1, "col":2, "CellOnFire":false}, ...]}`
- **Fields:**
  - `positions` (array): A list of cells where the fire has been extinguished.
