<script lang="ts">
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
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
					<ol>
						<li>
							Add <code>no-reply@saveto.ink</code> to
							<a
								href="https://www.amazon.com/gp/sendtokindle/email/"
								target="_blank"
								rel="external noopener">your Approved Personal Document E-mail List</a
							>
						</li>
						<li>
							<p>Save your Amazon Kindle email address to enable sending articles to your device</p>
						</li>
					</ol>
					<p></p>
				</legend>

				<input
					type="email"
					name="kindleEmail"
					required
					autocomplete="email"
					placeholder="abcd1234@kindle.com"
				/>
			</fieldset>
			<button type="submit">Enable Kindle delivey</button>
		</form>
	{/if}
</section>
