# Bomberman Game: A Beginner's Guide to the Architecture

Welcome! This document will walk you through how the Bomberman game works, explaining the concepts in a simple, easy-to-understand way. 

### The Two Big Pieces: Frontend and Backend

Imagine you're at a restaurant. You (the **customer**) sit at a table and read a menu. The **waiter** takes your order, and the **kitchen** prepares the food. You don't see how the kitchen works, you just get your meal. 

In our game, it's very similar:

1.  **The Frontend:** This is what you, the player, see and interact with in your web browser. It's like the **customer** and the **menu**. It shows you the game board, your character, and the buttons you can press.

2.  **The Backend:** This is the hidden engine of the game that runs on a server. It's like the **kitchen**. It manages all the game rules, keeps track of every player, and decides what happens when you drop a bomb.

These two parts talk to each other constantly over the internet.

---

## The Backend (The Game's "Brain")

The backend is written in a programming language called **Go**. It's the single source of truth for the game. If the backend says a player is at a certain spot, then that's where they are. This prevents cheating and ensures everyone sees the same game state.

#### What the Backend Does:

*   **Manages Player Connections:** When you join a game, the backend establishes a special, continuous connection with your browser called a **WebSocket**. Think of it like a walkie-talkie. It's always on, allowing for instant, two-way communication. This is how you get real-time updates.

*   **Handles Game Logic:** The backend enforces all the rules:
    *   Where players can move.
    *   How long it takes for a bomb to explode.
    *   What parts of the wall are destroyed in an explosion.
    *   Who wins or loses.

*   **Broadcasts Updates:** When one player moves, the backend doesn't just tell that one player. It **broadcasts** the move to *every other player* in the game. This is how you can see your opponents moving on your screen.

#### Key Backend Files:

*   `server.go`: This is the front door. It starts the server and listens for new players trying to connect.
*   `backend/bomberman/`: This folder is the core game engine.
    *   `Game.go`: The main file. It holds the entire game board, the list of players, bombs, and power-ups.
    *   `player.go`: A blueprint that defines what a "player" is (their name, lives, position, etc.).
    *   `bomb.go`: A blueprint for bombs. It handles the explosion timer and calculates which cells are affected.
    *   `move.go`: Handles player movement. It checks for collisions with walls, bombs, or other players.
    *   `broadcast.go`: The game's messenger. It takes updates (like a player moving) and sends them out to all connected players.

---

## The Frontend (What You See)

The frontend is what runs in your web browser. It's built with the three core technologies of the web:

*   **HTML:** The skeleton of the page.
*   **CSS:** The styling that makes the game look good (the colors, fonts, and layout).
*   **JavaScript:** The logic that makes the page interactive. It draws the game and communicates with the backend.

#### What the Frontend Does:

*   **Renders the Game:** It takes the data sent from the backend (like the grid layout and player coordinates) and visually draws it on the screen.

*   **Handles Your Input:** When you press the arrow keys to move or the spacebar to drop a bomb, the frontend captures this input.

*   **Sends Messages to the Backend:** After capturing your input, it sends a small message to the backend over the WebSocket (the "walkie-talkie"). For example, it sends a "start moving left" message.

*   **Receives Messages from the Backend:** It's always listening for messages from the backend. When it receives a message like "Player 2 moved to a new position," it immediately redraws Player 2 on your screen.

#### Key Frontend Files:

*   `index.html`: The main HTML file that your browser first loads.
*   `public/styles/`: This folder holds all the CSS files that define the game's appearance.
*   `public/images/`: This folder contains all the game's images, like the player sprites, bombs, and wall textures.
*   `app/`: This is where the main JavaScript logic lives.
    *   `main.js`: The starting point for the frontend. It sets up the connection to the backend.
    *   `game.js`: This is the most important file for gameplay. It handles drawing the game grid, players, and power-ups. It also listens for your keyboard presses.
    *   `start.js` & `lobby.js`: These files manage the start screen and the waiting lobby before the game begins.

---

## The Full Process Flow: A Player's Journey

Let's trace the entire process from start to finish.

1.  **Joining the Game:**
    *   You open the game's address in your browser.
    *   The frontend shows you a start screen (`start.js`).
    *   You enter your name and click "Join".
    *   The frontend sends your chosen name to the backend.
    *   The backend checks if the name is available. If it is, it creates a new player for you and sends a confirmation back.
    *   The frontend receives this confirmation and opens the WebSocket ("walkie-talkie") connection. You are now in the lobby.

2.  **Playing the Game:**
    *   When enough players have joined, the backend starts the game and sends the initial game state (the map, player positions) to everyone.
    *   The frontend (`game.js`) receives this data and draws the game on your screen.

3.  **Making a Move:**
    *   You press the **right arrow key**.
    *   Your frontend sends a tiny message to the backend: `{"msgType": "MS", "d": "r"}` (Move Start, direction right).
    *   The backend receives this. Its `move.go` logic updates your player's position.
    *   The backend then **broadcasts** a message to *everyone*: `{"MT":"M","PI":0,"XL":155,"YL":50,"D":"r"}` (Player 0 moved to X:155, Y:50, facing right).
    *   Your browser, and every other player's browser, receives this message and redraws your character in the new position.

This entire loop happens in a fraction of a second, creating the illusion of smooth, real-time movement.

4.  **Game Over:**
    *   When only one player is left, the backend's logic in `Game.go` detects this.
    *   It sends a final `GameState: GameOver` message to all players.
    *   Your frontend receives this and displays the "Game Over" screen.
