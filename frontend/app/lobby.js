
import { createElement } from '../framework/dom.js';
import { router } from '../framework/router.js';
import { APIUrl } from './main.js';

export default function Lobby() {
    let nameInput, joinButton, playerName, gameStatus, nameEntry, waitingRoom, countdownMessage, countdownSpan;

    const joinHandler = async (e) => {
        console.log('handling eveng')
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
                    alert('Strating the game')
                    router.navigate('/game')
                    break;
            }
        };

        ws.onclose = () => {
            console.log('WebSocket connection closed');
            gameStatus.dom.textContent = 'Connection lost';
        };

        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            gameStatus.dom.textContent = 'Connection error';
        };
    }

    function updatePlayerList(players) {
        const playerList = document.getElementById('player-list');
        playerList.innerHTML = '';
        players.forEach(player => {
            const li = document.createElement('li');
            li.textContent = player;
            playerList.appendChild(li);
        });
    }

    nameInput = createElement('input', { type: 'text', id: 'name-input', placeholder: 'Enter your name', onkeydown: joinHandler });
    joinButton = createElement('button', { id: 'join-game', onclick: joinHandler }, 'Join Game');
    playerName = createElement('span', { id: 'player-name' });
    gameStatus = createElement('span', { id: 'game-status' }, 'Not started');

    nameEntry = createElement('div', { id: 'name-entry' },
        nameInput,
        joinButton,
    );

    waitingRoom = createElement('div', { id: 'waiting-room', style: 'display: none;' },
        createElement('h2', {}, 'Welcome, ', playerName, '!'),
        createElement('p', {}, 'Waiting for other players to join...'),
        createElement('p', {}, 'Game Status: ', gameStatus),
        createElement('h3', {}, 'Players:'),
        createElement('ul', { id: 'player-list' })
    );

    return createElement('div', {},
        createElement('h1', {}, 'Bomberman Lobby'),
        nameEntry,
        waitingRoom,
    );
}
