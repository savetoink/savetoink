<script lang="ts">
	let { data } = $props();
	const kindleEmail = $derived(data?.kindle_email);
</script>

<section>
	<h2>Kindle delivery</h2>
	{#if kindleEmail}
		<p>Email delivery enabled for <code>{kindleEmail}</code></p>
		<form method="POST" action="?/deleteProfile">
			<button type="submit">Disable Kindle delivery</button>
		</form>
	{:else}
		<form method="POST" action="?/updateProfile">
			<fieldset>
				<legend>
					<p>Add your Amazon Kindle Email to enable sending articles to your device</p>
					<p>
						Remember to add first <code>no-reply@saveto.ink</code> to
						<a href="https://www.amazon.com/gp/sendtokindle/email/">your Amazon Kindle whitelist</a>
					</p>
				</legend>

				<input
					type="email"
					name="kindleEmail"
					required
					autocomplete="email"
					placeholder="yourname@kindle.com"
				/>
			</fieldset>
			<button type="submit">Enable Kindle delivey</button>
		</form>
	{/if}
</section>
