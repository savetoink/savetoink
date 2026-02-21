<script lang="ts">
	let { data } = $props();
</script>

<section>
	<h2>Kindle delivery</h2>
	<form method="POST" action="?/updateProfile">
		<fieldset>
			{#if data?.kindle_email}
				<legend>Email address associated with your Kindle device</legend>
			{:else}
				<legend>
					<p>Add your Amazon Kindle Email to enable sending articles to your device</p>
					<p>
						Remember to add first <code>no-reply@saveto.ink</code> to
						<a href="https://www.amazon.com/gp/sendtokindle/email/">your Amazon Kindle whitelist</a>
					</p>
				</legend>
			{/if}

			<input
				type="email"
				name="kindleEmail"
				required
				autocomplete="email"
				value={data?.kindle_email}
				disabled={data?.kindle_email}
				placeholder="yourname@kindle.com"
			/>
		</fieldset>
		{#if !data?.kindle_email}
			<button type="submit">Enable Kindle delivey</button>
		{/if}
	</form>

	{#if data?.kindle_email}
		<form method="POST" action="?/deleteProfile">
			<button type="submit">Disable Kindle delivery</button>
		</form>
	{/if}
</section>
