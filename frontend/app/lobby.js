
import { createElement, store } from '../framework/framework.js';

export default function Lobby() {
	const { players, countdown, playerId } = store.getState();

	const playerList = players.map(player => {
		const isYou = player.name === playerId;
		return createElement('div', { class: `player ${isYou ? 'you' : ''}` }, `${player.name} ${isYou ? '(You)' : ''}`);
	});

	return createElement('div', { class: 'lobby-container' },
		createElement('h1', {}, 'Lobby'),
		countdown !== null ? createElement('h2', {}, `Waiting for more players ${countdown}s`) : null,
		createElement('div', { id: 'player-list' }, ...playerList)
	);
}


