import { createElement } from '../framework/dom.js';
import { store } from '../framework/state.js';

export default function Game() {
    const render = () => {
        const state = store.getState();
        return createElement('div', {}, [
            createElement('h1', {}, 'Game Started!'),
            createElement('p', {}, `Game Status: ${state.gameStatus || 'Loading...'}`),
        ]);
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

    return createElement('div', { onMount: setup, onUnmount: teardown }, [
        render()
    ]);
}