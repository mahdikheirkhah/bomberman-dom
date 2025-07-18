import { createElement, store, router } from '../framework/framework.js';
import { APIUrl, handleWebSocket } from './main.js';

const joinHandler = async (e) => {
	if (e.key && e.key !== 'Enter') {
		return;
	}
	const name = document.getElementById('name-input').value;
	if (!name || name.trim().length === 0) {
		store.setState({ error: 'Please enter a name' });
		return;
	}

	try {
        // Step 1: Send an HTTP GET request to check name availability
        // Assuming your backend has an endpoint like /checkName that returns
        // a status indicating if the name is available or taken.
        const checkResponse = await fetch(`http://${APIUrl}/checkName?name=${encodeURIComponent(name)}`);
		const checkResult = await checkResponse.json(); // Assuming JSON response

        if (!checkResponse.ok) {
            const errorText = checkResult.reason
            store.setState({ error: `${errorText}`});
            return;
        }

        if (!checkResult.uuid) { // Assuming backend sends { isTaken: true/false }
            store.setState({ error: `Error creating player` });
            return;
        }

        // Step 2: If the name is available, proceed to establish WebSocket connection
        const ws = new WebSocket(`ws://${APIUrl}/ws?UUID=${encodeURIComponent(checkResult.uuid)}`);

        ws.onclose = () => {
            console.log('WebSocket connection closed.');
            // Handle disconnection, e.g., show a message, try to reconnect
            store.setState({ ws: null, gameStarted: false, gameData: null, error: 'Disconnected from game.' });
            router.navigate('/');
        };

		ws.onopen = () => {
		store.setState({ ws: ws, playerId: name }); // Using name as a temporary ID
		handleWebSocket();
		console.log('Websocket connection opened for player ', name)
		};

		ws.onerror = () => {
		console.log('Websocket connection error for player ', name)
		store.setState({ error: 'Connection error' });
		};

    } catch (error) {
        console.error('Failed to connect:', error);
        store.setState({ error: `Failed to connect to game: ${error.message}` });
    }
}



export default function Start() {
	const { error, playerId } = store.getState();

	const inputProps = {
		type: 'text',
		id: 'name-input',
		placeholder: 'Enter your name',
		onkeydown: joinHandler
	};

	if (playerId) {
		inputProps.value = playerId;
	}

	return createElement('div', { class: 'start-wrapper' },
		createElement('div', { class: 'bg-blur' }),
		createElement('div', { class: 'bg-main' },
			createElement('div', { class: 'start-container', style: '' },
				createElement('div', { class: 'start-form' },
					createElement('p', { class: error==='' ? 'hidden' : 'error-message' }, error),
					createElement('input', inputProps),
					createElement('button', { class: 'join-button', onclick: joinHandler}, 'Join Game')
				)
			)
		)
	);
}
