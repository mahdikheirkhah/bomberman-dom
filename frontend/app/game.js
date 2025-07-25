
import { createElement, store, router, patch, patchChildren } from '../framework/framework.js';
import { renderChat } from './chat.js';

// WebSocket message sender
const sendMsg = (msg) => {
    const { ws } = store.getState();
    if (ws) {
        ws.send(JSON.stringify(msg));
    }
};

const cellSize = 50

const powerupTypes = [
    { name: 'Extra Bomb', type: 'ExtraBomb', image: '/public/images/whiteegg.png', description: 'Increases bomb capacity by one.' },
    { name: 'Bomb Range', type: 'BombRange', image: '/public/images/extrab.webp', description: 'Increases bomb explosion range.' },
    { name: 'Extra Life', type: 'ExtraLife', image: '/public/images/life.webp', description: 'Grants an extra life.' },
    { name: 'Speed Boost', type: 'SpeedBoost', image: '/public/images/fast.webp', description: 'Increases movement speed.' }
];

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
            createElement('p', {}, player.lives > 0 ? '🩵'.repeat(player.lives) : 'Dead 💀')
        )
    );
}

// Render the game grid
function renderGameGrid(panel, players, powerups) {
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

                const defaultBackground = '#0098b8d1';
                const explosionBackground = `radial-gradient(circle at ${explosionX}px ${explosionY}px, #8e0404 0%, #0098b8d1 15%)`;

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

    const powerupElements = powerups.map(powerup => {
        const x = powerup.column * cellSize + cellSize; // Adjust for border
        const y = powerup.row * cellSize + cellSize;  // Adjust for border
        let powerUpImage;
        let additionalElement = null;

        const powerupType = powerupTypes.find(pt => pt.type === powerup.type);
        powerUpImage = powerupType ? powerupType.image : '';


        return createElement('div', { class: 'power-up', style: `transform: translate(${x}px, ${y}px);` },
            createElement('img', { src: powerUpImage, class: 'power-up-image' }),
            additionalElement
        );
    });

    return createElement('div', { class: 'game-grid' },
        ...borderedPanel.map(row =>
            createElement('div', { class: 'grid-row' },
                ...row.map(cell => {
                    if (cell === 'B') {
                        return createElement('div', { class: 'grid-cell' },
                            createElement('img', { src: '/public/images/redegg.png', class: 'bomb-image' })
                        );
                    } else if (cell === 'E') {
                        return createElement('div', { class: 'grid-cell E' });
                    } else {
                        return createElement('div', { class: `grid-cell ${cell}` });
                    }
                })
            )
        ),
        ...playerElements,
        ...powerupElements
    );
}

export function GameOverModal() {
    const { winner, gameData } = store.getState();
    const { players } = gameData;

    const playAgainHandler = () => {
        const { ws } = store.getState();
        if (ws) {
            ws.close();
        }
        // Keep playerId to pre-fill the name input
        store.setState({
            currentView: 'start',
            error: '',
            players: [],
            ws: null,
            countdown: null,
            gameStarted: false,
            gameOver: false,
            gameData: null,
            chatMessages: [],
            gameListenersAttached: false,
            playerAnimation: new Map(),
            powerups: [],
        });
        router.navigate('/');
    };

    return createElement('div', { class: 'modal' },
        createElement('div', { class: 'modal-content' },
            createElement('h2', {}, 'Game Over'),
            (winner >= 0 && players[winner]) ? createElement('p', {}, `${players[winner].name} wins!`) : createElement('p', {}, "It's a draw!"),
            createElement('button', { class: 'play-again-btn', onclick: playAgainHandler }, 'Play Again')
        )
    );
}


function gameLoop() {
    const { gameStarted, gameData, powerups } = store.getState();
    if (!gameStarted || !gameData) {
        return;
    }

    const { players, panel } = gameData;
    const mainGameArea = document.querySelector('.main-game-area');

    if (mainGameArea && mainGameArea.__vnode) {
        const newGrid = renderGameGrid(panel, players, powerups);
        patch(mainGameArea.__vnode, newGrid);
        mainGameArea.__vnode = newGrid;
    }

    requestAnimationFrame(gameLoop);
}

// Main Game component
export default function Game() {
    const { countdown, gameStarted, gameData, chatMessages, gameListenersAttached, powerups, gameOver } = store.getState();

    if (gameStarted && !gameListenersAttached) {
        setupEventListeners();
        store.setState({ gameListenersAttached: true });
        requestAnimationFrame(gameLoop);
    }

    if (!gameStarted || !gameData) {
        const countdownNumber = countdown > 10 ? 10 : countdown;

        const powerupElements = powerupTypes.map(powerup => {
            return createElement('div', { class: 'powerup-item' },
                createElement('img', { src: powerup.image, class: 'powerup-img' }),
                createElement('span', {}, `${powerup.name}: ${powerup.description}`)
            );
        });

        const modal = createElement('div', { id: 'instructions-modal', class: 'modal' },
            createElement('div', { class: 'modal-content' },
                createElement('h2', {}, 'How to Play'),
                createElement('p', {}, 'Use the arrow keys to move your penguin.'),
                createElement('p', {}, 'Press the spacebar to drop a bomb.'),
                createElement('h2', {}, 'Power-ups'),
                createElement('div', { id: 'powerups-container', class: 'powerups-container' }, ...powerupElements),
                renderChat(chatMessages || [])
            )
        );

        if (!window.penguinInterval) {
            let x = 0;
            let direction = 'right';
            window.penguinInterval = setInterval(() => {
                const penguin = document.querySelector('.penguin');
                if (penguin) {
                    penguin.classList.add('moving');

                    if (direction === 'right' && x >= 150) {
                        direction = 'left';
                    } else if (direction === 'left' && x <= 0) {
                        direction = 'right';
                    }

                    if (direction === 'right') {
                        x += 1;
                        penguin.classList.remove('face-l');
                        penguin.classList.add('face-r');
                    } else {
                        x -= 1;
                        penguin.classList.remove('face-r');
                        penguin.classList.add('face-l');
                    }

                    penguin.style.transform = `translateX(${x}px)`;
                }
            }, 20);
        }

        return createElement('div', { class: 'game-container game-starting-countdown-bg' },
            modal,
            createElement('img', { src: '/public/images/ice1.png', class: 'ice-image ice1' }),
            createElement('img', { src: '/public/images/ice2.png', class: 'ice-image ice2' }),
            createElement('div', { class: 'ice-image ice3-container' },
                createElement('img', { src: '/public/images/ice3.png', class: 'ice3-image' }),
                createElement('div', { class: 'screen-container' },
                    createElement('img', { src: '/public/images/screen.png', class: 'screen-image' }),
                    countdown !== null ? createElement('img', { src: `/public/images/${countdownNumber}.png`, class: 'countdown-number' }) : null
                )
            ),
            createElement('div', { class: 'ice-image ice4-container' },
                createElement('img', { src: '/public/images/ice4.png', class: 'ice4-image' }),
                createElement('div', { class: 'penguin face-r' })
            )
        );
    }

    // Request animation frame to ensure the grid is rendered before resizing
    requestAnimationFrame(handleResize);

    if (window.penguinInterval) {
        clearInterval(window.penguinInterval);
        window.penguinInterval = null;
    }

    const { players, panel } = gameData;

    let gameGridVnode;

    const mainGameArea = createElement('div', {
        class: 'main-game-area',
        ref: (element) => {
            if (element) {
                element.__vnode = gameGridVnode;
            }
        }
    }, (gameGridVnode = renderGameGrid(panel, players, powerups)));


    return createElement('div', { class: 'game-layout' },
        createElement('div', { class: 'player-panels' },
            ...players.map(renderPlayerPanel)
        ),
        mainGameArea,
        renderChat(chatMessages || []),
        gameOver ? GameOverModal(gameData) : null
    );
}
