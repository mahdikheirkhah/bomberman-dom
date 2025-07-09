# WebSocket Messages Documentation

This document outlines the WebSocket messages used in the Bomberman game, detailing their format, direction, and purpose.

## General Concepts

- **Client:** The frontend JavaScript application running in the user's browser.
- **Server:** The backend Go application.
- **Direction:** Indicates the flow of the message (e.g., Client -> Server, Server -> Client).
- **Format:** The JSON structure of the message.

---

## Client to Server Messages

These messages are sent from the client to the server.

### 1. Player Movement

- **Type:** `m`
- **Direction:** Client -> Server
- **Description:** Sent when the player presses a movement key (Arrow keys).
- **Format:**
  ```json
  {
    "msgType": "m",
    "d": "up" | "down" | "left" | "right"
  }
  ```

### 2. Place Bomb

- **Type:** `b`
- **Direction:** Client -> Server
- **Description:** Sent when the player presses the spacebar to place a bomb.
- **Format:**
  ```json
  {
    "msgType": "b"
  }
  ```

### 3. Chat Message

- **Type:** `c`
- **Direction:** Client -> Server
- **Description:** Sent when a player sends a message in the game chat.
- **Format:**
  ```json
  {
    "msgType": "c",
    "content": "Your message here"
  }
  ```

---

## Server to Client Messages

These messages are sent from the server to the client(s).

### 1. Game State Change

- **Type:** `GameState`
- **Direction:** Server -> Client
- **Description:** Informs the client about changes in the overall game state.
- **Format:**
  ```json
  {
    "type": "GameState",
    "state": "PlayerAccepted" | "LobbyCountdown" | "GameCountdown" | "GameStarted"
  }
  ```
  - **`PlayerAccepted`**: Sent to a player when they have successfully joined the game.
  - **`LobbyCountdown`**: Sent to all players when the minimum number of players has been reached, starting the lobby countdown.
  - **`GameCountdown`**: Sent to all players when the game is about to start.
  - **`GameStarted`**: Sent to all players when the game officially begins.

### 2. Player List Update

- **Type:** `player_list`
- **Direction:** Server -> Client
- **Description:** Sent to all clients whenever a new player joins, providing the updated list of all players.
- **Format:**
  ```json
  {
    "type": "player_list",
    "players": [
      {
        "name": "string",
        "lives": "int",
        "score": "int",
        "color": "string",
        "row": "int",
        "column": "int",
        "xlocation": "int",
        "yLocation": "int",
        "isDead": "bool",
        "numberOfBombs": "int",
        "numberOfUsedBombs": "int",
        "bombDelay": "int",
        "bombRange": "int",
        "stepSize": "int",
        "DirectionFace": "byte"
      }
    ]
  }
  ```

### 3. Lobby Countdown

- **Type:** `lobbyCountdown`
- **Direction:** Server -> Client
- **Description:** Provides the remaining seconds in the lobby countdown.
- **Format:**
  ```json
  {
    "type": "lobbyCountdown",
    "seconds": "int"
  }
  ```

### 4. Game Countdown

- **Type:** `gameCountdown`
- **Direction:** Server -> Client
- **Description:** Provides the remaining seconds before the game starts.
- **Format:**
  ```json
  {
    "type": "gameCountdown",
    "seconds": "int"
  }
  ```

### 5. Game Start

- **Type:** `gameStart`
- **Direction:** Server -> Client
- **Description:** Sent to all clients to signal the start of the game, including the initial game board layout and player positions.
- **Format:**
  ```json
  {
    "type": "gameStart",
    "players": "[see player_list format]",
    "numberOfPlayers": "int",
    "panel": "array[array[string]]"
  }
  ```

### 6. Chat Message

- **Type:** `CM`
- **Direction:** Server -> Client
- **Description:** Broadcasts a chat message from a player to all other players.
- **Format:**
  ```json
  {
    "type": "CM",
    "name": "string",
    "content": "string",
    "date": "timestamp",
    "filter": "bool",
    "senderIndex": "int",
    "color": "string"
  }
  ```

### 7. Move Accepted

- **Type:** `MA`
- **Direction:** Server -> Client
- **Description:** Confirms that a player's move was valid and provides their new position.
- **Format:**
  ```json
  {
    "type": "MA",
    "p": "string (player name)",
    "XL": "int (x location)",
    "YL": "int (y location)",
    "R": "int (row)",
    "C": "int (column)"
  }
  ```

### 8. Bomb Accepted

- **Type:** `BA`
- **Direction:** Server -> Client
- **Description:** Confirms that a bomb has been placed.
- **Format:**
  ```json
  {
    "MT": "BA",
    "XL": "int (x location)",
    "YL": "int (y location)",
    "R": "int (row)",
    "C": "int (column)"
  }
  ```

### 9. Bomb Not Accepted

- **Type:** `BNA`
- **Direction:** Server -> Client
- **Description:** Informs the player that they cannot place a bomb (e.g., they have no bombs left).
- **Format:**
  ```json
  {
    "MT": "BNA"
  }
  ```

### 10. Player Lives Decreased

- **Type:** `PLD`
- **Direction:** Server -> Client
- **Description:** Sent when a player is hit by an explosion and loses a life.
- **Format:**
  ```json
  {
    "type": "PLD",
    "lives": "int",
    "color": "string"
  }
  ```

### 11. Player Death

- **Type:** `PD`
- **Direction:** Server -> Client
- **Description:** Sent when a player has lost all their lives.
- **Format:**
  ```json
  {
    "type": "PD",
    "player": "[see player_list format]"
  }
  ```

### 12. Explosion Cells

- **Type:** `EXC`
- **Direction:** Server -> Client
- **Description:** Sent when a bomb explodes, indicating which cells are on fire.
- **Format:**
  ```json
  {
    "type": "EXC",
    "positions": [
      {
        "row": "int",
        "col": "int",
        "CellOnFire": "bool"
      }
    ]
  }
  ```

### 13. Turn Off Fire

- **Type:** `OF`
- **Direction:** Server -> Client
- **Description:** Sent when the fire from an explosion has extinguished.
- **Format:**
  ```json
  {
    "type": "OF",
    "positions": [
      {
        "row": "int",
        "col": "int"
      }
    ]
  }
  ```
