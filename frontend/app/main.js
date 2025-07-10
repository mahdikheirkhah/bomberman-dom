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
	gameData: null,
	chatMessages: [],
	gameListenersAttached: false, // Add this flag
});

export function handleWebSocket() {
	const { ws } = store.getState();
	if (!ws) {
		return;
	}

	ws.onmessage = (event) => {
		const message = JSON.parse(event.data);

        // The backend sends messages with either a 'type' or an 'MT' property.
        // We handle them accordingly.

		if (message.type) {
            switch (message.type) {
                case 'player_list':
                    store.setState({ players: message.players });
                    break;
                case 'GameState':
                    if (message.state === 'LobbyCountdown') {
                        store.setState({ countdown: null, gameStarted: false });
                    } else if (message.state === 'GameCountdown') {
                        store.setState({ currentView: 'game', gameStarted: false });
                        router.navigate("/game");
                    } else if (message.state === 'GameStarted') {
                        store.setState({ countdown: null, gameStarted: true });
                    } else if (message.state === 'PlayerAccepted') {
                        store.setState({ currentView: 'lobby', playerId: message.playerId });
                    }
                    break;
                case 'gameStart':
                    store.setState({ gameData: { players: message.players, panel: message.panel } });
                    break;
                case 'lobbyCountdown':
                case 'gameCountdown':
                    store.setState({ countdown: message.seconds });
                    break;
                case 'CM':
                    const { chatMessages } = store.getState();
                    store.setState({ chatMessages: [...chatMessages, { player: message.name, message: message.content }] });
                    break;
                case 'playerUpdate':
                    store.setState({ gameData: { ...store.getState().gameData, players: message.players, panel: message.panel } });
                    break;
                case 'bombUpdate':
                case 'explosion':
                    store.setState({ gameData: { ...store.getState().gameData, panel: message.panel } });
                    break;
                case 'playerDead':
                    store.setState({ gameData: { ...store.getState().gameData, players: message.players } });
                    break;
                case 'gameOver':
                    store.setState({ currentView: 'start', gameStarted: false, gameData: null, countdown: null, players: [], chatMessages: [] });
                    router.navigate('/');
                    break;
            }
        }

        if (message.MT) {
            switch (message.MT) {
                case 'M': // Player Move
                    const { gameData } = store.getState();
                    if (gameData && gameData.players) {
                        const updatedPlayers = gameData.players.map(player => {
                            if (player.index === message.PI) {
                                return { ...player, xlocation: message.XL, yLocation: message.YL, DirectionFace: message.D };
                            }
                            return player;
                        });
                        store.setState({ gameData: { ...gameData, players: updatedPlayers } });
                    }
                    break;
                case 'BA':
                    const { gameData: gameDataBomb } = store.getState();
                    if (gameDataBomb && gameDataBomb.panel) {
                        const newPanel = [...gameDataBomb.panel];
                        newPanel[message.R][message.C] = 'B';
                        store.setState({ gameData: { ...gameDataBomb, panel: newPanel } });
                    }
                    break;
            }
        }
	};

	ws.onclose = (event) => {
		console.log('Websocket connection closed for player ', name)
		if (event.code === 1008) {
			store.setState({ error: 'Game is full' });
		} else {
			store.setState({ error: 'Connection lost' });
		}
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