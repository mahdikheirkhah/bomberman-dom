import { createElement, store, router } from '../framework/framework.js';
import { APIUrl, handleWebSocket } from './main.js';

const joinHandler = async (e) => {
	if (e.key && e.key !== 'Enter') {
		return;
	}
	const name = document.getElementById('name-input').value;
	if (!name) {
		store.setState({ ...store.getState(), error: 'Please enter a name' });
		return;
	}

	const ws = new WebSocket(`ws://${APIUrl}/ws?name=${name}`);

	ws.onopen = () => {
		const state = store.getState();
		store.setState({ ...state, ws: ws, playerId: name }); // Using name as a temporary ID
		handleWebSocket();
		router.navigate('/lobby');
	};

	ws.onclose = (event) => {
		const state = store.getState();
		if (event.code === 4000) {
			store.setState({ ...state, error: 'Game is full' });
		} else {
			store.setState({ ...state, error: 'Connection failed' });
		}
	};

	ws.onerror = () => {
		store.setState({ ...store.getState(), error: 'Connection error' });
	};
};

export default function Start() {
	const { error } = store.getState();

	return createElement('div', { class: 'start-container' },
		createElement('h1', {}, 'Bomberman'),
		createElement('input', {
			type: 'text',
			id: 'name-input',
			placeholder: 'Enter your name',
			onkeydown: joinHandler
		}),
		createElement('button', { onclick: joinHandler }, 'Join Game'),
		error ? createElement('p', { class: 'error' }, error) : null
	);
}
