
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
    const chatInput = document.querySelector('.chat-input-form input');
    if (chatInput && document.activeElement === chatInput) {
        return;
    }

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
    window.addEventListener('resize', handleResize);
}

function handleResize() {
    const gameArea = document.querySelector('.main-game-area');
    const gameGrid = document.querySelector('.game-grid');

    if (!gameArea || !gameGrid) return;

    const { gameData } = store.getState();
    if (!gameData || !gameData.panel || gameData.panel.length === 0) return;

    const numCols = gameData.panel[0].length + 2; // +2 for borders
    const numRows = gameData.panel.length + 2;   // +2 for borders
    const gridWidth = numCols * cellSize;
    const gridHeight = numRows * cellSize;

    const areaWidth = gameArea.clientWidth;
    const areaHeight = gameArea.clientHeight;

    const scale = Math.min(areaWidth / gridWidth, areaHeight / gridHeight);

    gameGrid.style.transform = `scale(${scale})`;
}

// Render a single player panel
function renderPlayerPanel(player) {
    const avatarClass = `player-avatar ${player.color}${player.lives <= 0 ? ' dead' : ''}`;
    return createElement('div', { class: 'player-card' },
        createElement('div', { class: avatarClass }),
        createElement('div', { class: 'player-info' },
            createElement('h3', {}, player.name),
            createElement('p', {}, player.lives > 0 ? '❤️'.repeat(player.lives) : 'Dead 💀')
        )
    );
}

// Render the game grid
function renderGameGrid(panel, players) {
    const { playerAnimation } = store.getState();
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

    const playerElements = players.filter(player => !player.isDead).map(player => {
        const x = player.xlocation + cellSize; // Adjust for border
        const y = player.yLocation + cellSize;  // Adjust for border
        const animation = playerAnimation.get(player.index) || { isMoving: false, isHurt: false };

        const playerClasses = [
            'player',
            animation.isMoving ? 'moving' : 'stopped',
            animation.isHurt ? 'hurt' : ''
        ].join(' ');

        if (animation.isHurt) {
            const gameGrid = document.querySelector('.game-grid');
            if (gameGrid) {
                const playerSize = 48; // from css
                const explosionX = x + playerSize / 2;
                const explosionY = y + playerSize / 2;

                const defaultBackground = 'radial-gradient(circle, #199a9ed1 0%, #004878 100%)';
                const explosionBackground = `radial-gradient(circle at ${explosionX}px ${explosionY}px, #8e0404 0%, #199a9ed1 50%, #004878 100%)`;

                gameGrid.style.background = explosionBackground;
                gameGrid.classList.add('explosion');

                setTimeout(() => {
                    gameGrid.style.background = defaultBackground;
                    gameGrid.classList.remove('explosion');
                }, 500); // Match animation duration
            }
        }

        const spriteClasses = [
            'player-sprite',
            player.color,
            `face-${player.DirectionFace}`
        ].join(' ');

        return createElement('div', {
            class: playerClasses,
            style: `transform: translate(${x}px, ${y}px);`
        }, createElement('div', { class: spriteClasses }));
    });

    return createElement('div', { class: 'game-grid' },
        ...borderedPanel.map(row =>
            createElement('div', { class: 'grid-row' },
                ...row.map(cell => {
                    if (cell === 'B') {
                        return createElement('div', { class: 'grid-cell' },
                            createElement('img', { src: '/public/bomb.svg', class: 'bomb-image' })
                        );
                    } else if (cell === 'E') {
                        return createElement('div', { class: 'grid-cell E' });
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
    const { playerId } = store.getState();

    const handleSubmit = (e) => {
        e.preventDefault();
        const input = e.target.elements.message;
        if (input.value) {
            sendMsg({ msgType: 'c', content: input.value });
            input.value = '';
        }
    };

    const renderMessage = (msg) => {
        const { playerIndex } = store.getState();
        const isSent = msg.senderIndex === playerIndex;
        const bubbleClass = isSent ? 'message-bubble sent' : 'message-bubble received';
        const sender = isSent ? 'You' : msg.player;
        const timestamp = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false });

        return createElement('div', { class: 'chat-message' },
            createElement('div', { class: bubbleClass },
                createElement('div', { class: 'message-sender', style: `color: ${msg.color}` }, sender),
                createElement('div', { class: 'message-content' }, msg.message),
                createElement('div', { class: 'message-timestamp' }, timestamp)
            )
        );
    };

    return createElement('div', { class: 'game-chat' },
        createElement('div', { class: 'resize-handle', onmousedown: onMouseDown }),
        createElement('div', { class: 'chat-header' }, 'Game Chat'),
        createElement('div', { class: 'chat-messages' },
            ...messages.map(renderMessage)
        ),
        createElement('form', { class: 'chat-input-form', onsubmit: handleSubmit },
            createElement('input', { type: 'text', name: 'message', placeholder: 'Type a message...' }),
            createElement('button', { type: 'submit' }, '➤')
        )
    );
}

function onMouseDown(e) {
    e.preventDefault();
    document.addEventListener('mousemove', onMouseMove);
    document.addEventListener('mouseup', onMouseUp);
}

function onMouseMove(e) {
    const chat = document.querySelector('.game-chat');
    if (chat) {
        const newWidth = window.innerWidth - e.clientX;
        chat.style.width = `${newWidth}px`;
    }
}

function onMouseUp() {
    document.removeEventListener('mousemove', onMouseMove);
    document.removeEventListener('mouseup', onMouseUp);
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

    // Request animation frame to ensure the grid is rendered before resizing
    requestAnimationFrame(handleResize);

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
