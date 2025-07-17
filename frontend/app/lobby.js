import { createElement, store } from '../framework/framework.js';

export default function Lobby() {
    const { players, countdown, playerId } = store.getState();

    const playerList = players.map(player => {
        const isYou = player.name === playerId;
        return createElement('div', { class: `${isYou ? 'you' : ''}` }, `${player.name} ${isYou ? '(You)' : ''}`);
    });

    const playersJoined = players.length;

    return createElement('div', { class: 'countdown-bg' },
        createElement('link', { rel: 'stylesheet', href: '/app/countdown.css' }),
        createElement('link', { rel: 'stylesheet', href: '/app/modal.css' }),
        createElement('img', { src: '/public/ice1.png', class: 'ice-image ice1' }),
        createElement('img', { src: '/public/ice2.png', class: 'ice-image ice2' }),
        createElement('div', { class: 'ice3-container' },
            createElement('img', { src: '/public/ice3.png', class: 'ice3-image' }),
        ),
        createElement('div', { class: 'ice4-container' },
            createElement('img', { src: '/public/ice4.png', class: 'ice4-image' })
        ),
        createElement('div', { id: 'lobby-modal', class: 'modal' },
            createElement('div', { class: 'modal-content' },
                createElement('h2', {}, 'Waiting for more players'),
                createElement('h3', {}, `Current amount of players: ${playersJoined}`),
                countdown !== null ? createElement('h3', {}, `Waiting for more players ${countdown}s`) : null,
                createElement('div', { id: 'player-list' }, ...playerList)
            )
        )
    );
}