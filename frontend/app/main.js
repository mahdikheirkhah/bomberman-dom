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
	playerId: null,
	ws: null,
	countdown: null,
	gameStarted: false,
});

export function handleWebSocket() {
	const { ws } = store.getState();
	if (!ws) {
		console.log('No WS!!!')
		return;
	}

	ws.onmessage = (event) => {
		const message = JSON.parse(event.data);

		switch (message.type) {
			case 'player_list':
				store.setState({ players: message.players });
				break;
			case 'GameState':
				if (message.state === 'LobbyCountdown') {
					store.setState({ countdown: null, gameStarted: false });
				} else if (message.state === 'GameCountdown') {
					store.setState({ currentView: 'game', gameStarted: false });
				} else if (message.state === 'GameStarted') {
					store.setState({ countdown: null, gameStarted: true });
				}
				break;
			case 'lobbyCountdown':
				store.setState({ countdown: message.seconds });
				break;
			case 'gameCountdown':
				store.setState({ countdown: message.seconds });
				break;
		}
	};

	ws.onclose = () => {
		store.setState({ error: 'Connection lost' });
	};

	ws.onerror = () => {
		store.setState({ error: 'Connection error' });
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
router.addRoute('/', () => store.setState({ currentView: 'start' }));
router.addRoute('/lobby', () => store.setState({ currentView: 'lobby' }));
router.addRoute('/game', () => store.setState({ currentView: 'game' }));
router.setDefaultHandler(() => store.setState({ currentView: 'start' }));

createApp(App, document.getElementById('app'));

export { APIUrl };
