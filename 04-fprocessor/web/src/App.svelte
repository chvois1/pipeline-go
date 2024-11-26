<script>
	import Router, {location, link} from 'svelte-spa-router';
	import Home from './Home.svelte';
	import Dashboard from './Dashboard.svelte';
	import Monitor from './Monitor.svelte';
	import Subscribe from './Subscribe.svelte';
	import Log from './Log.svelte';
	import About from './About.svelte';
</script>

<nav>
	<a href="/">Home</a>
	<a href="/#/dashboard">Dashboard</a>
	<a href="/#/subscribe">Subscribe</a>
	<a href="/#/monitor">Monitor</a>
	<a href="/#/log">Log</a>
	<a href="/#/about">About</a>
</nav>

<Router routes={{
	'/': Home,
	'/dashboard': Dashboard,
	'/subscribe': Subscribe,
	'/monitor': Monitor,
	'/log': Log,
	'/about': About
}} />
<style>
	nav {
		border: 1px;
		background-color: indigo;
		color: white;
		padding: 6px;
	}

	a {
		text-decoration: none;
		color: white;
		font-size: 20px;
	}
</style>