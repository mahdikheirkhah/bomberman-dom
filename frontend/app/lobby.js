
import { createElement, store } from '../framework/framework.js';

export default function Lobby() {
	const { players, countdown, playerId } = store.getState();

	const playerList = players.map(player => {
		const isYou = player.name === playerId;
		return createElement('div', { class: `${isYou ? 'you' : ''}` }, `${player.name} ${isYou ? '(You)' : ''}`);
	});

	return createElement('div', { class: 'lobby-container' },
		createElement('h1', {}, 'Lobby'),
		countdown !== null ? createElement('h2', {}, `Waiting for more players ${countdown}s`) : null,
		createElement('div', { id: 'player-list' }, ...playerList)
	);
}
const chatForm = document.getElementById('chat-form');
const chatMessages = document.getElementById('chat-messages');
const chatInput = document.getElementById('chat-input');

chatForm.addEventListener('submit', function(e) {
  e.preventDefault();

  const message = chatInput.value.trim();
  if (message === '') return;

  // Create message element
  const messageElement = document.createElement('div');
  messageElement.textContent = message;
  messageElement.classList.add('chat-message');

  chatMessages.appendChild(messageElement);
  chatMessages.scrollTop = chatMessages.scrollHeight;

  chatInput.value = '';
});


