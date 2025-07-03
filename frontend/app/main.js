import { createApp } from '../framework/app.js';
import { router } from '../framework/router.js';
import { store } from '../framework/state.js';
import Lobby from './lobby.js';
import Game from './game.js';

const APIUrl = 'localhost:8080'

// Initialize game status in the store
store.setState({ gameStatus: 'waiting', currentView: Lobby });

// Function to render a component by updating the store
const renderComponent = (component) => {
    store.setState({ currentView: component });
};

// Add routes
router.addRoute('/', () => renderComponent(Lobby));
router.addRoute('/game', () => renderComponent(Game));

// Set default route handler
router.setDefaultHandler(() => renderComponent(Lobby));


function App() {
    const state = store.getState();
    return state.currentView();
}

createApp(App, document.getElementById('app'));

export { APIUrl }