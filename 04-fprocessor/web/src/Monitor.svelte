<script>
	import { items, time } from './storemon.js';

	const formatter = new Intl.DateTimeFormat('en', {
		hour12: true,
		hour: 'numeric',
		minute: '2-digit',
		second: '2-digit'
	});

  let obj = { id: "FileReceiver", cmd: "stop", status:"off"}
  // {FileReceiver, DataIntegrity, DataDispatcher, FileMaker, DataArchiving, FileSender}
  let result = null
  async function doPost (item) {
    console.log("=>doPost")
    console.log("item:" + item.description)
    console.log("obj:" + JSON.stringify(obj))
		const res = await fetch('http://localhost:8080/api/v1/states/current', {
			method: "PUT",
			body: JSON.stringify(obj)
		})
		
		const json = await res.json()
    console.log("json:" + json)
		result = JSON.stringify(json)
    console.log("result:" + result)
    console.log("<=doPost")
	}

</script>
<div class="todoapp stack-large">
  <p> 
    {formatter.format($time)}
  </p>
  <hr>
  <ul  class="todo-list stack-large" aria-labelledby="list-heading">
  {#each $items as item}
  <li class="todo">
    <label>
      <!--
<button style="color:white;background-color:hsl({item.state}, 100%, 50%)">{item.description}</button>
      -->
      {item.description}
      <button type="button" on:click={doPost(item)}>x</button>
    </label>
  </li>
  {/each}
</ul>
</div>
