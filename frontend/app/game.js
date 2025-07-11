
import { createElement, store } from '../framework/framework.js';

// WebSocket message sender
const sendMsg = (msg) => {
    const { ws } = store.getState();
    if (ws) {
        ws.send(JSON.stringify(msg));
    }
};

const cellSize = 50

// Set to track currently pressed movement keys
const pressedKeys = new Set();

// Player movement and bomb placement
const handleKeyEvent = (e, isKeyDown) => {
    if (e.repeat) return;

    const moveKeys = ['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight'];
    const key = e.code;

    if (moveKeys.includes(key)) {
        if (isKeyDown) {
            sendMsg({ msgType: 'MS', d: getDirection(key) });
            pressedKeys.add(key);
        } else {
            pressedKeys.delete(key);
            // If all movement keys are released, send ME
            if (pressedKeys.size === 0) {
                sendMsg({ msgType: 'ME' });
            }
        }
    } else if (key === 'Space' && isKeyDown) {
        sendMsg({ msgType: 'b' }); // Only send bomb on keydown
    }
};

// Helper to get direction from key code
const getDirection = (key) => {
    switch (key) {
        case 'ArrowUp': return 'u';
        case 'ArrowDown': return 'd';
        case 'ArrowLeft': return 'l';
        case 'ArrowRight': return 'r';
        default: return '';
    }
};

// Add and remove event listeners
function setupEventListeners() {
    document.addEventListener('keydown', (e) => handleKeyEvent(e, true));
    document.addEventListener('keyup', (e) => handleKeyEvent(e, false));
}

// Render a single player panel
function renderPlayerPanel(player) {
    return createElement('div', { class: 'player-card' },
        createElement('div', { class: `player-avatar ${player.color}` }),
        createElement('div', { class: 'player-info' },
            createElement('h3', {}, player.name),
            createElement('p', {}, `Bombs: ${player.numberOfBombs}`),
            createElement('p', {}, `Lives: ${player.lives}`)
        )
    );
}

// Render the game grid
function renderGameGrid(panel, players) {
    const borderedPanel = [];
    const numRows = panel.length + 2;
    const numCols = panel[0].length + 2;

    for (let i = 0; i < numRows; i++) {
        borderedPanel[i] = [];
        for (let j = 0; j < numCols; j++) {
            if (i === 0 || i === numRows - 1 || j === 0 || j === numCols - 1) {
                borderedPanel[i][j] = 'W';
            } else {
                borderedPanel[i][j] = panel[i - 1][j - 1];
            }
        }
    }

    const playerElements = players.map(player => {
        const left = player.xlocation + cellSize; // Adjust for center coordinates and border
        const top = player.yLocation + cellSize;  // Adjust for center coordinates and border

        return createElement('div', {
            class: `player ${player.color}`,
            style: `left: ${left}px; top: ${top}px;`
        });
    });

    return createElement('div', { class: 'game-grid' },
        ...borderedPanel.map(row =>
            createElement('div', { class: 'grid-row' },
                ...row.map(cell => {
                    if (cell === 'B') {
                        return createElement('div', { class: 'grid-cell' },
                            createElement('img', { src: '/public/bomb.svg', class: 'bomb-image' })
                        );
                    } else {
                        return createElement('div', { class: `grid-cell ${cell}` });
                    }
                })
            )
        ),
        ...playerElements
    );
}

// Render the chat area
function renderChat(messages) {
    const handleSubmit = (e) => {
        e.preventDefault();
        const input = e.target.elements.message;
        if (input.value) {
            sendMsg({ msgType: 'c', content: input.value });
            input.value = '';
        }
    };

    return createElement('div', { class: 'game-chat' },
        createElement('h3', {}, 'Chat'),
        createElement('div', { class: 'chat-messages' },
            ...messages.map(msg => createElement('p', {}, createElement('b', {}, `${msg.player}: `), msg.message))
        ),
        createElement('form', { onsubmit: handleSubmit },
            createElement('input', { type: 'text', name: 'message', placeholder: 'Type a message...' }),
            createElement('button', { type: 'submit' }, 'Send')
        )
    );
}

// Main Game component
export default function Game() {
    const { countdown, gameStarted, gameData, chatMessages, gameListenersAttached } = store.getState();

    if (gameStarted && !gameListenersAttached) {
        setupEventListeners();
        store.setState({ gameListenersAttached: true });
    }

    if (!gameStarted || !gameData) {
        return createElement('div', { class: 'game-container' },
            createElement('h1', {}, 'Bmbrmn'),
            countdown !== null ? createElement('h2', {}, `Game starting in ${countdown}s`) : null,
            gameStarted ? createElement('h1', {}, 'Game in Progress') : null
        );
    }

    const { players, panel } = gameData;

    return createElement('div', { class: 'game-layout' },
        createElement('div', { class: 'player-panels' },
            ...players.map(renderPlayerPanel)
        ),
        createElement('div', { class: 'main-game-area' },
            renderGameGrid(panel, players)
        ),
        renderChat(chatMessages || [])
    );
}
