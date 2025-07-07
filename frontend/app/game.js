import { createElement, store } from '../framework/framework.js';

export default function Game() {
	const { countdown } = store.getState();

	return createElement('div', { class: 'game-container' },
		createElement('h1', {}, 'Game in Progress'),
		countdown !== null ? createElement('h2', {}, `Game starting in ${countdown}s`) : null
	);
}
