import { createElement } from '../framework/dom.js';
import { router } from '../framework/router.js';
import { APIUrl } from './main.js';

let currentPlayerId = null; // To highlight "you"

export default function Lobby() {
    let nameInput, joinButton, playerName, gameStatus, nameEntry, waitingRoom, playerListContainer;

    const joinHandler = async (e) => {
        if (e.key === 'Enter' || e.type === 'click') {
            const name = nameInput.dom.value;
            if (name) {
                try {
                    const response = await fetch(`http://${APIUrl}/api/join`, {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ name }),
                    });

                    if (response.ok) {
                        const data = await response.json();
                        currentPlayerId = data.playerId;

                        playerName.dom.textContent = name;
                        nameEntry.dom.style.display = 'none';
                        waitingRoom.dom.style.display = 'block';
                        setupWebSocket(data.playerId);
                    } else {
                        const err = await response.json();
                        alert(`Error: ${err.error}`);
                    }
                } catch (error) {
                    console.error('Error joining game:', error);
                    alert('Failed to join the game. Please try again.');
                }
            }
        }
    };

    function setupWebSocket(playerId) {
        const ws = new WebSocket(`ws://${APIUrl}/ws?playerId=${playerId}`);

        ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            switch (message.type) {
                case 'game_status':
                    gameStatus.dom.textContent = message.payload.status;
                    updatePlayerList(message.payload.players);
                    break;
                case 'game_start':
                    router.navigate('/game');
                    break;
            }
        };

        ws.onclose = () => {
            console.log('WebSocket closed');
            gameStatus.dom.textContent = 'Connection lost';
        };

        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            gameStatus.dom.textContent = 'Connection error';
        };
    }
const countdownTimer = createElement('div', {
    id: 'countdown',
    style: 'display: none; font-size: 24px; color: white; margin-top: 10px;'
}, 'Starting in: ', createElement('span', { id: 'countdown-timer' }, '5'), 's');

let countdownInterval = null;
let countdownSeconds = 5;

function startCountdown() {
    const countdownEl = document.getElementById('countdown');
    const timerEl = document.getElementById('countdown-timer');

    if (countdownInterval) return;

    countdownSeconds = 5;
    countdownEl.style.display = 'block';
    timerEl.textContent = countdownSeconds;

    countdownInterval = setInterval(() => {
        countdownSeconds--;
        timerEl.textContent = countdownSeconds;

        if (countdownSeconds <= 0) {
            clearInterval(countdownInterval);
            countdownInterval = null;
            router.navigate('/game'); // or trigger game start
        }
    }, 1000);
}

function stopCountdown() {
    if (countdownInterval) {
        clearInterval(countdownInterval);
        countdownInterval = null;
        const countdownEl = document.getElementById('countdown');
        if (countdownEl) countdownEl.style.display = 'none';
    }
}
// // function toggleReady(playerId) {
//     fetch(`http://${APIUrl}/api/toggle-ready`, {
//         method: 'POST',
//         headers: { 'Content-Type': 'application/json' },
//         body: JSON.stringify({ playerId })
//     });
// // }
const statusDot = createElement('span', {
  class: `status-indicator ${player.ready ? 'status-ready' : 'status-waiting'}`
});

const playerDiv = createElement('div', {
  class: `player-card ${isYou ? 'you' : ''}`
}, avatar, name, statusDot);

function updatePlayerList(players) {
    playerListContainer.dom.innerHTML = '';

    let readyCount = 0;

    players.forEach((player, index) => {
        const isHost = index === 0;
        const isYou = player.id === currentPlayerId;
        const isReady = player.ready; // assuming your WebSocket message includes this field
        if (isReady) readyCount++;

        const avatar = createElement('img', {
            src: `public/assets/penguin${index + 1}.png`,
            class: 'player-avatar'
        });

        const name = createElement('span', {
            class: 'player-name'
        }, `${player.name}${isHost ? ' 👑' : ''}${isYou ? ' (You)' : ''}`);

        const statusDot = createElement('span', {
            class: `status-indicator ${isReady ? 'status-ready' : 'status-waiting'}`
        });

        const playerDiv = createElement('div', {
            class: `player-card ${isYou ? 'you' : ''}`,
            onclick: isYou ? () => toggleReady(player.id) : null
        }, avatar, name, statusDot);

        playerListContainer.dom.appendChild(playerDiv.dom || playerDiv);
    });

    if (readyCount >= 2 && players.find(p => p.id === currentPlayerId)?.ready) {
        startCountdown(); // you'll define this
    } else {
        stopCountdown();
    }
}

// Leave lobby button
const leaveButton = createElement('button', {
    id: 'leave-lobby',
    onclick: () => {
        location.reload(); // or implement a route
        }
    }, 'Leave Lobby');

    nameInput = createElement('input', {
        type: 'text',
        id: 'name-input',
        placeholder: 'Enter your name',
        onkeydown: joinHandler,
        'data-testid': 'name-input'
    });

    joinButton = createElement('button', {
        id: 'join-game',
        onclick: joinHandler,
        'data-testid': 'join-button'
    }, 'Join Game');

    playerName = createElement('span', { id: 'player-name' });
    gameStatus = createElement('span', { id: 'game-status' }, 'Waiting');

    nameEntry = createElement('div', { id: 'name-entry' },
        nameInput,
        joinButton
    );

    playerListContainer = createElement('div', { id: 'player-list', 'data-testid': 'player-list' });

    waitingRoom = createElement('div', { id: 'waiting-room', style: 'display: none;' },
        createElement('h2', {}, 'Welcome, ', playerName, '!'),
        createElement('p', {}, 'Game Status: ', gameStatus),
        createElement('h3', {}, 'Players in the lobby:'),
        playerListContainer,
        countdownTimer,
        leaveButton
    );

    return createElement('div', {},
        createElement('h1', {}, 'Bomberman Lobby'),
        nameEntry,
        waitingRoom
    );
}
