
import { createElement, store } from '../framework/framework.js';

// WebSocket message sender
const sendMsg = (msg) => {
    const { ws } = store.getState();
    if (ws) {
        ws.send(JSON.stringify(msg));
    }
};

// Render the chat area
export function renderChat(messages) {
    const { playerId } = store.getState();

    const handleSubmit = (e) => {
        e.preventDefault();
        const input = e.target.elements.message;
        if (input.value) {
            sendMsg({ msgType: 'c', content: input.value });
            input.value = '';
        }
    };

    const renderMessage = (msg) => {
        const { playerIndex } = store.getState();
        const isSent = msg.senderIndex === playerIndex;
        const bubbleClass = isSent ? 'message-bubble sent' : 'message-bubble received';
        const sender = isSent ? 'You' : msg.player;
        const timestamp = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false });

        return createElement('div', { class: 'chat-message' },
            createElement('div', { class: bubbleClass },
                createElement('div', { class: 'message-sender', style: `color: ${msg.color}` }, sender),
                createElement('div', { class: 'message-content' }, msg.message),
                createElement('div', { class: 'message-timestamp' }, timestamp)
            )
        );
    };

    return createElement('div', { class: 'game-chat' },
        createElement('div', { class: 'resize-handle', onmousedown: onMouseDown }),
        createElement('div', { class: 'chat-header' }, 'Game Chat'),
                createElement('div', {
            class: 'chat-messages',
            ref: (el) => {
                if (el) {
                    el.scrollTop = el.scrollHeight;
                }
            }
        },
            ...messages.map(renderMessage)
        ),
        createElement('form', { class: 'chat-input-form', onsubmit: handleSubmit },
            createElement('input', { type: 'text', name: 'message', placeholder: 'Type a message...' }),
            createElement('button', { type: 'submit' }, '➤')
        )
    );
}

function onMouseDown(e) {
    e.preventDefault();
    document.addEventListener('mousemove', onMouseMove);
    document.addEventListener('mouseup', onMouseUp);
}

function onMouseMove(e) {
    const chat = document.querySelector('.game-chat');
    if (chat) {
        const newWidth = window.innerWidth - e.clientX;
        chat.style.width = `${newWidth}px`;
    }
}

function onMouseUp() {
    document.removeEventListener('mousemove', onMouseMove);
    document.removeEventListener('mouseup', onMouseUp);
}
