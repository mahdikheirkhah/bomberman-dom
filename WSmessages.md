# Bomberman WebSocket API Documentation (Corrected)

This document outlines the WebSocket messages used for communication between the game client (frontend) and the game server (backend), based on the current implementation.

## General Concepts

- **Client-to-Server (C->S):** Messages sent from the player's browser to the server.
- **Server-to-Client (S->C):** Messages sent from the server to one or more players' browsers.
- Most messages are JSON objects.
- There are two primary formats for server-to-client messages, identified by either a `type` field or an `MT` (MessageType) field.

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

#### `player_list`
- **Description:** Provides the current list of players in the lobby.
- **Payload:** `{"type":"player_list","players":[...]}`

#### `GameState`
- **Description:** Informs the client of a major change in the game's state.
- **Payload:** `{"type":"GameState","state":"LobbyCountdown"}`
- **Possible States:** `PlayerAccepted`, `LobbyCountdown`, `GameCountdown`, `GameStarted`.

#### `lobbyCountdown` / `gameCountdown`
- **Description:** Provides the remaining seconds in a countdown.
- **Payload:** `{"type":"lobbyCountdown","seconds":5}`

#### `gameStart`
- **Description:** Sent when the game begins, containing the initial board and player data.
- **Payload:** `{"type":"gameStart","players":[...],"panel":[...]}`

#### `CM` (Chat Message)
- **Description:** Broadcasts a chat message to all players.
- **Payload:** `{"type":"CM","name":"player1","content":"Hi!","color":"G"}`

#### `playerUpdate`
- **Description:** Sent to update player and panel data simultaneously.
- **Payload:** `{"type":"playerUpdate","players":[...],"panel":[...]}`

#### `bombUpdate` / `explosion`
- **Description:** Sent when bombs or explosions change the game grid. The frontend uses both `bombUpdate` and `explosion` cases to handle updates to the game panel.
- **Payload:** `{"type":"...","panel":[...]}`

#### `playerDead`
- **Description:** Sent when a player has lost all their lives.
- **Payload:** `{"type":"playerDead","players":[...]}`

#### `gameOver`
- **Description:** Sent when the game has ended.
- **Payload:** `{"type":"gameOver"}`

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
- **Description:** Confirms a bomb has been placed and provides its grid location.
- **Payload:** `{"MT":"BA","R":2,"C":4}`
- **Fields:**
  - `R` (number): Row index.
  - `C` (number): Column index.