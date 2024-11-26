import { writable } from 'svelte/store';

let msgs = [
	{
		"id":          1,
		"state":       100,
		"description": "File receiver",
	},
	{
		"id":          2,
		"state":       100,
		"description": "Data integrity",
	},
	{
		"id":          3,
		"state":       100,
		"description": "Data dispatcher",
	},
	{
		"id":          4,
		"state":       100,
		"description": "File maker",
	},
	{
		"id":          5,
		"state":       100,
		"description": "Data archiving",
	},
	{
		"id":          6,
		"state":       100,
		"description": "File sender",
	}
]

export const items = writable([]);
const socket = new WebSocket('ws://localhost:8080/api/v1/mon');

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

export const time = writable(new Date(), function start(set) {
	const interval = setInterval(() => {
		set(new Date());
	}, 1000);

	return function stop() {
		clearInterval(interval);
	};
});
