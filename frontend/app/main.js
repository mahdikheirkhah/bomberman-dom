import { createApp, store, router } from '../framework/framework.js';
import Start from './start.js';
import Lobby from './lobby.js';
import Game from './game.js';

const APIUrl = `${window.location.hostname}:8080`;

// Initialize store
store.setState({
    currentView: 'start',
    error: '',
    players: [],
    playerId: null,
    ws: null,
    countdown: null,
    gameStarted: false,
    gameOver: false,
    gameData: null,
    chatMessages: [],
    gameListenersAttached: false, // Add this flag
    playerAnimation: new Map(), // For client-side animation
    powerups: [], // Add this line
});

const playerMoveTimers = new Map();

export function handleWebSocket() {
    const { ws } = store.getState();
    if (!ws) {
        return;
    }

    ws.onmessage = (event) => {
        const message = JSON.parse(event.data);
        const { gameData, playerAnimation, chatMessages } = store.getState();

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
                    } else if (message.state === 'GameOver') {
                        console.log('Game over received')
                        store.setState({ gameOver: true, winner: message.winner });
                    }
                    break;
                case 'PlayerAccepted':
                    store.setState({ currentView: 'lobby', playerIndex: message.index });
                    break;
                case 'PlayerDisconnected':
                    const playersList = store.getState().players;
                    store.setState({ players: playersList.filter(player => player.index !== message.index) });
                    if (playersList.length - 1 < 2) {
                        //store.removeState('countdown');
                        store.setState({ countdown: null });
                    }
                    break;
                case 'StopCountdown':
                    store.setState({ countdown: null });
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
                    store.setState({ chatMessages: [...chatMessages, { player: message.name, message: message.content, senderIndex: message.senderIndex, color: message.color }] });
                    break;
                case 'playerUpdate':
                    store.setState({ gameData: { ...store.getState().gameData, players: message.players, panel: message.panel } });
                    break;
                case 'bombUpdate':
                case 'explosion':
                    store.setState({ gameData: { ...store.getState().gameData, panel: message.panel } });
                    break;
                case 'AddPowerup':
                    store.setState({ powerups: [...store.getState().powerups, message.powerup] });
                    break;
                case 'RemovePowerup':
                    store.setState({ powerups: store.getState().powerups.filter(p => p.row !== message.row || p.column !== message.column) });
                    break;
                case 'EatBombPowerup':
                    if (gameData && gameData.players) {
                        const updatedPlayers = gameData.players.map(p => {
                            if (p.index === message.player) {
                                return { ...p, NumberOfBombs: message.numberOfBombs, NumberOfUsedBombs: message.numberOfUsedBombs };
                            }
                            return p;
                        });
                        store.setState({ gameData: { ...gameData, players: updatedPlayers } });
                    }
                    break;
                case 'EatLifePowerup':
                    if (gameData && gameData.players) {
                        const updatedPlayers = gameData.players.map(p => {
                            if (p.index === message.player) {
                                return { ...p, lives: message.numberOfLives };
                            }
                            return p;
                        });
                        store.setState({ gameData: { ...gameData, players: updatedPlayers } });
                    }
                    break;
                case 'PD':
                    if (gameData && gameData.players) {
                        const updatedPlayers = gameData.players.map(p => {
                            if (p.index === message.player.index) {
                                return { ...message.player, lives: 0 }; // Update player data and ensure lives are 0
                            }
                            return p;
                        });
                        store.setState({ gameData: { ...gameData, players: updatedPlayers } });
                    }
                    break;
                case 'PLD':
                    if (gameData && gameData.players) {
                        const updatedPlayers = gameData.players.map(p => {
                            if (p.index === message.playerIndex) {
                                return { ...p, lives: message.lives };
                            }
                            return p;
                        });

                        const animationState = playerAnimation.get(message.playerIndex) || { isMoving: false };
                        animationState.isHurt = true;
                        playerAnimation.set(message.playerIndex, animationState);

                        setTimeout(() => {
                            animationState.isHurt = false;
                            playerAnimation.set(message.playerIndex, animationState);
                            store.setState({ playerAnimation: new Map(playerAnimation) });
                        }, 500); // Animation duration

                        store.setState({ gameData: { ...gameData, players: updatedPlayers }, playerAnimation: new Map(playerAnimation) });
                    }
                    break;

                case 'PR': // Player Respawn
                    if (gameData && gameData.players) {
                        const updatedPlayers = gameData.players.map(p => {
                            if (p.index === message.playerIndex) {
                                return { ...p, xlocation: message.xlocation, yLocation: message.yLocation };
                            }
                            return p;
                        });
                        store.setState({ gameData: { ...gameData, players: updatedPlayers } });
                    }
                    break;
            }
        }

        if (message.MT) {
            switch (message.MT) {
                case 'M': // Player Move
                    if (gameData && gameData.players) {
                        const updatedPlayers = gameData.players.map(player => {
                            if (player.index === message.PI) {
                                return { ...player, xlocation: message.XL, yLocation: message.YL, DirectionFace: message.D };
                            }
                            return player;
                        });

                        // Animation logic
                        const animationState = playerAnimation.get(message.PI) || { isMoving: false };
                        animationState.isMoving = true;
                        playerAnimation.set(message.PI, animationState);

                        // Reset timer to stop animation
                        if (playerMoveTimers.has(message.PI)) {
                            clearTimeout(playerMoveTimers.get(message.PI));
                        }
                        playerMoveTimers.set(message.PI, setTimeout(() => {
                            animationState.isMoving = false;
                            playerAnimation.set(message.PI, animationState);
                            store.setState({ playerAnimation: new Map(playerAnimation) });
                        }, 150));

                        store.setState({
                            gameData: { ...gameData, players: updatedPlayers },
                            playerAnimation: new Map(playerAnimation)
                        });
                    }
                    break;
                case 'BA':
                    if (gameData && gameData.panel) {
                        const newPanel = [...gameData.panel];
                        newPanel[message.R][message.C] = 'B';
                        store.setState({ gameData: { ...gameData, panel: newPanel } });
                    }
                    break;
                case 'EXC':
                    console.log('Explosion message received:', message);
                    if (gameData && gameData.panel) {
                        const newPanel = [...gameData.panel];
                        newPanel[message.bombRow][message.bombCol] = 'E';

                        message.positions.forEach(pos => {
                            newPanel[pos.row][pos.col] = 'E';
                        });
                        store.setState({ gameData: { ...gameData, panel: newPanel } });
                    }
                    break;
                case 'OF':
                    if (gameData && gameData.panel) {
                        const newPanel = [...gameData.panel];
                        message.positions.forEach(pos => {
                            newPanel[pos.row][pos.col] = '';
                        });
                        store.setState({ gameData: { ...gameData, panel: newPanel } });
                    }
                    break;
            }
        }
    };

    ws.onclose = (event) => {
        console.log('Websocket connection closed for player ')
        if (event.code === 1008) {
            store.setState({ error: event.reason });
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

const { ws } = store.getState();
const path = router.getCurrentPath();

if (!ws && path && path !== '/') {
    router.navigate('/');
} else {
    router.init();
}

export { APIUrl };