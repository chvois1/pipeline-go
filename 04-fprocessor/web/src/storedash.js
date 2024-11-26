import { writable } from 'svelte/store';

let msgs = []
export const items = writable([]);
const socket = new WebSocket('ws://localhost:8080/api/v1/ws');

// Connection opened
socket.addEventListener('open', function (event) {
    console.log("Web socket is opened");
});

// Listen for messages
socket.addEventListener('message', function (event) {
    console.log("Msg received: " + event.data);
	let jsonMsg = JSON.parse(event.data)
	let idx = msgs.findIndex(x => x.id === jsonMsg.id)
	0 > idx ? msgs.push(jsonMsg) : msgs[idx]=jsonMsg
	items.set(msgs)  
});