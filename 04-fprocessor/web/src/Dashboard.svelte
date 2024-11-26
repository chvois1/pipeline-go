<script>
    import { items } from './storedash.js';

    const StateEnum = Object.freeze({"doRecv":1, "doIntegrity":2, "doDispatcher" :3, "doFileMaker":4, "doArchiver":5, "doSend":6})

	let uid = $items.length + 1;

	function add(input) {
		const item = {
			id: uid++,
			state: StateEnum.doRecv,
			description: input.value
		};

		$items = [item, ...$items];
		input.value = '';
	}

	function remove(item) {
		$items = $items.filter(t => t !== item);
	}
</script>

<div class='board'>
	<input
		class="new-item"
		placeholder="Enter a new file name ..."
		on:keydown="{event => event.key === 'Enter' && add(event.target)}"
	>

	<div class='source'>
		<h3>source</h3>
		<p>File receiver</p>
		{#each $items.filter(t => t.state == StateEnum.doRecv ) as item (item.id)}
			<label>
            	{item.description}
				<button on:click="{() => item.state=StateEnum.doIntegrity}">x</button>
			</label>
		{/each}
	</div>
	<div class='stage'>
		<h3>stage</h3>
		<p>Data integrity</p>
		{#each $items.filter(t => t.state == StateEnum.doIntegrity) as item (item.id)}
			<label>
				{item.description}
				<button on:click="{() => item.state=StateEnum.doDispatcher}">x</button>
			</label>
		{/each}
	</div>
	<div class='stage'>
		<h3>stage</h3>
		<p>Data dispatcher</p>
		{#each $items.filter(t => t.state == StateEnum.doDispatcher) as item (item.id)}
			<label>
				{item.description}
				<button on:click="{() => item.state=StateEnum.doFileMaker}">x</button>
			</label>
		{/each}
	</div>
	<div class='stage'>
		<h3>stage</h3>
		<p>File maker</p>
		{#each $items.filter(t => t.state == StateEnum.doFileMaker) as item (item.id)}
			<label>
				{item.description}
				<button on:click="{() => item.state=StateEnum.doArchiver}">x</button>
			</label>
		{/each}
	</div>
	<div class='stage'>
		<h3>stage</h3>
		<p>Data archiving</p>
		{#each $items.filter(t => t.state == StateEnum.doArchiver) as item (item.id)}
			<label>
				{item.description}
				<button on:click="{() => item.state=StateEnum.doSend}">x</button>
			</label>
		{/each}
	</div>
	<div class='sink'>
		<h3>sink</h3>
		<p>File sender</p>
		{#each $items.filter(t => t.state == StateEnum.doSend) as item (item.id)}
			<label>
				{item.description}
				<button on:click="{() => remove(item)}">x</button>
			</label>
		{/each}
	</div>
</div>

<style>
	.new-item {
		font-size: 1.4em;
		width: 100%;
		margin: 2em 0 1em 0;
	}

	.board {
		max-width: 100%;
		margin: 0 auto;
	}

	.source, .stage, .sink {
		float: left;
		width: 15%;
		padding: 0 1em 0 0;
		box-sizing: border-box;
	}

	label {
		top: 0;
		left: 0;
		display: block;
		font-size: 1em;
		line-height: 1;
		padding: 0.5em;
		margin: 0 auto 0.5em auto;
		border-radius: 2px;
		background-color: #eee;
		user-select: none;
	}

	input { margin: 0 }

    .source label {
		background-color:red;
	}
	.stage label {
		background-color: orange;
	}

	.sink label {
		background-color: green;
	}

	button {
		float: right;
		height: 1em;
		box-sizing: border-box;
		padding: 0 0.5em;
		line-height: 1;
		background-color: transparent;
		border: none;
		color: rgb(170,30,30);
		opacity: 0;
		transition: opacity 0.2s;
	}

	label:hover button {
		opacity: 1;
	}
</style>