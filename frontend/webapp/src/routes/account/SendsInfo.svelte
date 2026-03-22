<script lang="ts">
	import type { SendsResponse, SendsResponseNoLimits } from '@savetoink/shared';

	let { sends }: { sends: SendsResponse | SendsResponseNoLimits | undefined } = $props();

	const hasQuotaInfo = $derived(
		sends && 'current_sends' in sends && 'max_sends_per_period' in sends
	);
	const totalSends = $derived(sends?.total_sends ?? 0);
</script>

<section>
	<h2>Article sends</h2>

	{#if sends}
		<p>Total sends (all time): <strong>{totalSends}</strong></p>

		{#if hasQuotaInfo && 'current_sends' in sends}
			{@const quota = sends as SendsResponse}
			<article>
				<header>
					<h3>Quota information</h3>
				</header>
				<dl>
					<dt>Sends in current period</dt>
					<dd>{quota.current_sends} / {quota.max_sends_per_period}</dd>

					<dt>Remaining sends</dt>
					<dd>{quota.remaining_sends}</dd>

					<dt>Period length</dt>
					<dd>{quota.period_days} days</dd>
				</dl>
			</article>
		{/if}
	{:else}
		<p>Unable to fetch sends information.</p>
	{/if}
</section>
