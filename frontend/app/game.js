
import { createElement, store } from '../framework/framework.js';

// WebSocket message sender
const sendMsg = (msg) => {
    const { ws } = store.getState();
    if (ws) {
        ws.send(JSON.stringify(msg));
    }
};

// Player movement and bomb placement
const handleKeyDown = (e) => {
    if (e.repeat) return;
    switch (e.code) {
        case 'ArrowUp':
            sendMsg({ msgType: 'm', d: 'up' });
            break;
        case 'ArrowDown':
            sendMsg({ msgType: 'm', d: 'down' });
            break;
        case 'ArrowLeft':
            sendMsg({ msgType: 'm', d: 'left' });
            break;
        case 'ArrowRight':
            sendMsg({ msgType: 'm', d: 'right' });
            break;
        case 'Space':
            sendMsg({ msgType: 'b' });
            break;
    }
};

// Add and remove event listeners
function setupEventListeners() {
    document.addEventListener('keydown', handleKeyDown);
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
    const playerElements = players.map(player => {
        return createElement('div', {
            class: `player ${player.color}`,
            style: `left: ${player.xlocation}px; top: ${player.yLocation}px;`
        });
    });

    return createElement('div', { class: 'game-grid' },
        ...panel.map(row =>
            createElement('div', { class: 'grid-row' },
                ...row.map(cell => createElement('div', { class: `grid-cell ${cell}` }))
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
