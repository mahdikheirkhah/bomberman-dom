import { createElement, store } from '../framework/framework.js';
import { renderChat } from './chat.js';

export default function Lobby() {
    const { players, countdown, playerId, chatMessages } = store.getState();

    const colorToImage = {
        "R": "/public/images/R.gif",
        "Y": "/public/images/Y.gif",
        "G": "/public/images/G.gif",
        "B": "/public/images/B.gif",
    };

    const playerList = players.map(player => {
        const isYou = player.name === playerId;

        const playerImage = createElement('img', {
            src: colorToImage[player.color],
            alt: player.color,
            style: 'width: 48px; height: 48px; border-radius: 50%; border: 2px solid white; margin-right: 10px; background-color: white'
        });

        const playerName = createElement('span', {}, `${player.name} ${isYou ? '(You)' : ''}`);

        return createElement(
            'div',
            {
                class: `player-list-item ${isYou ? 'you' : ''}`,
                style: 'display: flex; align-items: center; margin-bottom: 8px;'
            },
            playerImage,
            playerName
        );
    });

    const playersJoined = players.length;

    return createElement('div', { class: 'countdown-bg' },
        createElement('div', { id: 'lobby-modal', class: 'modal' },
            createElement('div', { class: 'modal-content' },
                createElement('h2', {}, 'Waiting for more players'),
                createElement('h3', {}, `Current amount of players: ${playersJoined}`),
                countdown !== null ? createElement('h3', {}, `Waiting for more players ${countdown}s`) : null,
                createElement('div', { id: 'player-list' }, ...playerList),
                renderChat(chatMessages || [])
            )
        )
    );

    // return createElement('div', { class: 'countdown-bg' },
    //     createElement('img', { src: '/public/images/ice1.png', class: 'ice-image ice1' }),
    //     createElement('img', { src: '/public/images/ice2.png', class: 'ice-image ice2' }),
    //     createElement('div', { class: 'ice3-container' },
    //         createElement('img', { src: '/public/images/ice3.png', class: 'ice3-image' }),
    //     ),
    //     createElement('div', { class: 'ice4-container' },
    //         createElement('img', { src: '/public/images/ice4.png', class: 'ice4-image' })
    //     ),
    //     createElement('div', { id: 'lobby-modal', class: 'modal' },
    //         createElement('div', { class: 'modal-content' },
    //             createElement('h2', {}, 'Waiting for more players'),
    //             createElement('h3', {}, `Current amount of players: ${playersJoined}`),
    //             countdown !== null ? createElement('h3', {}, `Waiting for more players ${countdown}s`) : null,
    //             createElement('div', { id: 'player-list' }, ...playerList)
    //         )
    //     ),
    //     renderChat(chatMessages || [])
    // );


}