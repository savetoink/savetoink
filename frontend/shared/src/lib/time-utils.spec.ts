import { describe, expect, it } from 'vitest';

describe('formatInstantFromEpochMs', () => {
	it('should format epoch milliseconds to a human-readable date string', async () => {
		const { formatInstantFromEpochMs } = await import('./time-utils.js');
		const epoch = 1_743_000_000_000;

		const result = formatInstantFromEpochMs(epoch);

		expect(result).toBeTypeOf('string');
		expect(result).toMatch(/^[A-Z][a-z]{2}, \d{2} [A-Z][a-z]{2} \d{4}, \d{2}:\d{2}:\d{2} [A-Z]{3,4}[-+]?\d{1,2}$/);
	});
});

describe('getCurrentDevDate', () => {
	it('should return current date in DDMMHHMM format', async () => {
		const { getCurrentDevDate } = await import('./time-utils.js');
		const result = getCurrentDevDate();

		expect(result).toBeTypeOf('string');
		expect(result).toMatch(/^\d{8}$/);
	});
});

describe('getCurrentYear', () => {
	it('should return current year as number', async () => {
		const { getCurrentYear } = await import('./time-utils.js');
		const result = getCurrentYear();

		expect(result).toBeTypeOf('number');
		expect(result).toBeGreaterThanOrEqual(2000);
		expect(result).toBeLessThanOrEqual(2100);
	});
});

describe('formatTimeRemainingFromEpochMs', () => {
	it('should format time remaining with hours, minutes, and seconds', async () => {
		const { formatTimeRemainingFromEpochMs } = await import('./time-utils.js');
		const futureTime = Date.now() + 5 * 60 * 60 * 1000 + 23 * 60 * 1000 + 45 * 1000;

		const result = formatTimeRemainingFromEpochMs(futureTime);

		expect(result).toBeTypeOf('string');
		expect(result).toMatch(/^\d+h \d+m \d+s$/);
	});

	it('should format time remaining with only minutes and seconds', async () => {
		const { formatTimeRemainingFromEpochMs } = await import('./time-utils.js');
		const futureTime = Date.now() + 23 * 60 * 1000 + 45 * 1000;

		const result = formatTimeRemainingFromEpochMs(futureTime);

		expect(result).toBeTypeOf('string');
		expect(result).toMatch(/^\d+m \d+s$/);
	});

	it('should format time remaining with only seconds', async () => {
		const { formatTimeRemainingFromEpochMs } = await import('./time-utils.js');
		const futureTime = Date.now() + 45 * 1000;

		const result = formatTimeRemainingFromEpochMs(futureTime);

		expect(result).toBeTypeOf('string');
		expect(result).toMatch(/^\d+s$/);
	});

	it('should return zero time for past timestamps', async () => {
		const { formatTimeRemainingFromEpochMs } = await import('./time-utils.js');
		const pastTime = Date.now() - 1000;

		const result = formatTimeRemainingFromEpochMs(pastTime);

		expect(result).toBe('0h 0m 0s');
	});

	it('should handle exact hours', async () => {
		const { formatTimeRemainingFromEpochMs } = await import('./time-utils.js');
		const futureTime = Date.now() + 2 * 60 * 60 * 1000;

		const result = formatTimeRemainingFromEpochMs(futureTime);

		expect(result).toMatch(/^2h \d+m \d+s$/);
	});

	it('should handle exact minutes', async () => {
		const { formatTimeRemainingFromEpochMs } = await import('./time-utils.js');
		const futureTime = Date.now() + 30 * 60 * 1000;

		const result = formatTimeRemainingFromEpochMs(futureTime);

		expect(result).toMatch(/^30m \d+s$/);
	});
});
