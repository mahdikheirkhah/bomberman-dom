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

    function updatePlayerList(players) {
        playerListContainer.dom.innerHTML = '';

        players.forEach((player, index) => {
            const isHost = index === 0;
            const isYou = player.id === currentPlayerId;

            let label = isHost ? '👑 Host' : '🧍 Player';
            if (isYou) label += ' (You)';

            const playerDiv = createElement('div', {
                class: `player ${isHost ? 'host' : ''}`,
                'data-testid': `player-${player.id}`
            }, `${label}: ${player.name}`);

            playerListContainer.dom.appendChild(playerDiv.dom || playerDiv);
        });

        if (players.length >= 2 && players[0].id === currentPlayerId) {
            const startBtn = createElement('button', {
                id: 'start-game',
                onclick: () => {
                    fetch(`http://${APIUrl}/api/start`, { method: 'POST' });
                }
            }, 'Start Game');
            playerListContainer.dom.appendChild(startBtn.dom || startBtn);
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
        leaveButton
    );

    return createElement('div', {},
        createElement('h1', {}, 'Bomberman Lobby'),
        nameEntry,
        waitingRoom
    );
}
