
# API and WebSocket Requirements

## HTTP API

### Endpoint: `POST /api/join`

This endpoint is used to register a player in the game lobby.

**Request Body:**

```json
{
  "name": "string"
}
```

**Success Response (200 OK):**

```json
{
  "playerId": "string",
  "message": "Welcome to the lobby!"
}
```

**Error Response (400 Bad Request):**

- If the name is empty or invalid.
- If the game is already full.

```json
{
  "error": "string"
}
```

## WebSocket Communication

The WebSocket connection is established after a successful response from the `/api/join` endpoint. The client will connect to `/ws`.

### Server-to-Client Messages

1.  **Game Status Update:**
    -   Sent to all clients in the lobby when a new player joins or the game status changes.
    -   **Message Format:**
        ```json
        {
          "type": "game_status",
          "payload": {
            "status": "waiting_for_players" | "game_in_progress" | "game_over",
            "players": ["player1", "player2", "player3"]
          }
        }
        ```

2.  **Game Start:**
    -   Sent to all clients when the game starts.
    -   **Message Format:**
        ```json
        {
          "type": "game_start"
        }
        ```

### Client-to-Server Messages

At this stage, the client only listens for server messages after the initial handshake. No client-to-server messages are required for the lobby functionality.
