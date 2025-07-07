import { createElement, store } from '../framework/framework.js';

export default function Game() {
	console.log('game funct running')
	const { countdown, gameStarted } = store.getState();

	return createElement('div', { class: 'game-container' },
		createElement('h1', {}, 'Bmbrmn'),
		countdown !== null ? createElement('h2', {}, `Game starting in ${countdown}s`) : null,
		gameStarted ? createElement('h1', {}, 'Game in Progress') : null
	);
}
