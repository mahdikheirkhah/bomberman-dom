# Bomberman DOM

This is a multiplayer real-time Bomberman game built with a Go backend and a vanilla JavaScript frontend. The game supports up to 4 players and features classic Bomberman gameplay with power-ups, destructible walls, and a chat system.

## Features

- **Real-time Multiplayer:** Play with up to 4 players in real-time.
- **Classic Bomberman Gameplay:** Place bombs, destroy walls, and defeat your opponents.
- **Power-ups:** Collect power-ups to increase your bomb count, bomb range, and movement speed.
- **In-Game Chat:** Communicate with other players using the in-game chat.
- **Dynamic Game Lobbies:** Join a lobby and wait for other players to start the game.

## Technologies Used

- **Backend:** Go
  - `gorilla/websocket` for WebSocket communication.
- **Frontend:** Vanilla JavaScript (ES6 Modules)
  - A lightweight custom framework for DOM manipulation and state management.

## Project Structure

```
/bomberman-dom
├── backend/
│   ├── bomberman/  # Game logic
│   ├── server.go   # Main server entrypoint
│   └── go.mod
├── frontend/
│   ├── app/        # Application-specific JavaScript
│   ├── framework/  # Custom frontend framework
│   ├── public/     # Static assets (images, styles)
│   └── index.html  # Main HTML file
└── README.md
```

## Getting Started

### Prerequisites

- [Go](https://golang.org/doc/install) (version 1.20 or higher)

### Setup and Running

1.  **Clone the repository:**

    ```bash
    git clone https://01.gritlab.ax/git/obalandi/bomberman-dom.git
    cd bomberman-dom
    ```

2.  **Run the backend server:**

    Open a terminal and run the following command from the project's root directory:

    ```bash
    go run ./backend/server.go
    ```

    The server will start on port `8080`.

3.  **Run the frontend:**

    Since the frontend is built with vanilla JavaScript and doesn't have any build steps, you can serve the `frontend` directory using any simple HTTP server. One of the easiest ways is to use Python's built-in HTTP server.

    Open a second terminal, navigate to the `frontend` directory, and run:

    ```bash
    # For Python 3
    python3 -m http.server 8000
    ```

    This will serve the frontend on port `8000`.

4.  **Play the game:**

    Open your web browser and navigate to `http://localhost:8000`. You can open multiple tabs to simulate multiple players.

## Contributors

- [Oleg Balandin](https://github.com/olegamobile)
- [Mohammad Mahdi Kheirkhah](https://github.com/mahdikheirkhah)
- [Inka Säävuori](https://github.com/Inkasaa)
- [Kateryna Ovsiienko](https://github.com/mavka1207)
- [Fatemeh Kheirkhah](https://github.com/fatemekh78)