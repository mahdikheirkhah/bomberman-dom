import { createApp, store, router } from '../framework/framework.js';
import Start from './start.js';
import Lobby from './lobby.js';
import Game from './game.js';

const APIUrl = 'localhost:8080';

// Initialize store
store.setState({
	currentView: 'start',
	error: '',
	players: [],
	countdown: null,
	playerId: null,
	ws: null
});

function handleWebSocket() {
	const { ws } = store.getState();
	if (!ws) return;

	ws.onmessage = (event) => {
		const message = JSON.parse(event.data);
		const state = store.getState();

		switch (message.type) {
			case 'player_list':
				store.setState({ ...state, players: message.players });
				break;
			case 'countdown':
				store.setState({ ...state, countdown: message.seconds });
				break;
			case 'GameState':
				if (message.state === 'GameCountdown') {
					router.navigate('/game');
				}
				break;
		}
	};

	ws.onclose = () => {
		store.setState({ ...store.getState(), error: 'Connection lost' });
	};

	ws.onerror = () => {
		store.setState({ ...store.getState(), error: 'Connection error' });
	};
}

function App() {
	const { currentView } = store.getState();
	switch (currentView) {
		case 'start':
			return Start();
		case 'lobby':
			return Lobby();
		case 'game':
			return Game();
		default:
			return Start();
	}
}

// Add routes that change the currentView
router.addRoute('/', () => store.setState(prevState => ({ ...prevState, currentView: 'start' })));
router.addRoute('/lobby', () => store.setState(prevState => ({ ...prevState, currentView: 'lobby' })));
router.addRoute('/game', () => store.setState(prevState => ({ ...prevState, currentView: 'game' })));
router.setDefaultHandler(() => store.setState(prevState => ({ ...prevState, currentView: 'start' })));

createApp(App, document.getElementById('app'));

export { APIUrl, handleWebSocket };
