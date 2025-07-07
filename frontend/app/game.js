import { createElement, store } from '../framework/framework.js';

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

function renderGameGrid(panel) {
	return createElement('div', { class: 'game-grid' },
		...panel.map(row =>
			createElement('div', { class: 'grid-row' },
				...row.map(cell => createElement('div', { class: `grid-cell ${cell}` }))
			)
		)
	);
}

function renderChat(messages) {
	return createElement('div', { class: 'game-chat' },
		createElement('h3', {}, 'Chat'),
		createElement('div', { class: 'chat-messages' },
			...messages.map(msg => createElement('p', {}, createElement('b', {}, `${msg.player}: `), msg.message))
		),
		createElement('input', { type: 'text', placeholder: 'Type a message...' })
	);
}

export default function Game() {
	const { countdown, gameStarted, gameData, chatMessages } = store.getState();

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
			renderGameGrid(panel)
		),
		renderChat(chatMessages || [])
	);
}
