<script lang="ts">
	import { resolve } from '$app/paths';
</script>

<nav>
	<ul>
		<li><a href={resolve('/')}>My List</a></li>
		<li><a href={resolve('/new')}>New</a></li>
		<li><a href={resolve('/login')}>Login</a></li>
	</ul>
</nav>
