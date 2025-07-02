import { createApp } from '../framework/app.js';
import { router } from '../framework/router.js';
import { store } from '../framework/state.js';
import Lobby from './lobby.js';
import Game from './game.js';

const APIUrl = 'localhost:8080'

// Initialize game status in the store
store.setState({ gameStatus: 'waiting', currentView: Lobby });
console.log('State after initial set:', store.getState());

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
    const render = () => {
        const state = store.getState();
        return state.currentView();
    };

    let unsubscribe;
    const setup = (element) => {
        unsubscribe = store.subscribe(() => {
            element.replaceWith(render());
        });
    };

    const teardown = () => {
        if (unsubscribe) {
            unsubscribe();
        }
    };

    return createApp(() => render(), document.getElementById('app'), setup, teardown);
}

App();
//router.init();

export { APIUrl }