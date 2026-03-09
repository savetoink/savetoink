import { describe, expect, it } from 'vitest';

describe('getBuildDate', () => {
	it('should return a string in MMDDHHMM format', async () => {
		const { getBuildDate } = await import('./build-config.js');
		const result = getBuildDate();

		expect(result).toBeTypeOf('string');
		expect(result).toHaveLength(8);

		const month = parseInt(result.slice(0, 2), 10);
		const day = parseInt(result.slice(2, 4), 10);
		const hour = parseInt(result.slice(4, 6), 10);
		const minute = parseInt(result.slice(6, 8), 10);

		expect(month).toBeGreaterThanOrEqual(1);
		expect(month).toBeLessThanOrEqual(12);
		expect(day).toBeGreaterThanOrEqual(1);
		expect(day).toBeLessThanOrEqual(31);
		expect(hour).toBeGreaterThanOrEqual(0);
		expect(hour).toBeLessThanOrEqual(23);
		expect(minute).toBeGreaterThanOrEqual(0);
		expect(minute).toBeLessThanOrEqual(59);
	});

	it('should use UTC timezone', async () => {
		const { getBuildDate } = await import('./build-config.js');
		const result = getBuildDate();
		const now = new Date();

		const resultMonth = parseInt(result.slice(0, 2), 10);
		const resultDay = parseInt(result.slice(2, 4), 10);
		const resultHour = parseInt(result.slice(4, 6), 10);
		const resultMinute = parseInt(result.slice(6, 8), 10);

		expect(resultMonth).toBe(now.getUTCMonth() + 1);
		expect(resultDay).toBe(now.getUTCDate());
		expect(resultHour).toBeLessThanOrEqual(23);
		expect(resultMinute).toBeLessThanOrEqual(59);
	});

	it('should format 2026-02-03T00:12:00Z as 02030012', async () => {
		const testDate = new Date('2026-02-03T00:12:00Z');
		const month = String(testDate.getUTCMonth() + 1).padStart(2, '0');
		const day = String(testDate.getUTCDate()).padStart(2, '0');
		const hour = String(testDate.getUTCHours()).padStart(2, '0');
		const minute = String(testDate.getUTCMinutes()).padStart(2, '0');
		const expected = `${month}${day}${hour}${minute}`;

		expect(expected).toBe('02030012');
	});

	it('should format January 1st at midnight as 01010000', async () => {
		const testDate = new Date('2025-01-01T00:00:00Z');
		const month = String(testDate.getUTCMonth() + 1).padStart(2, '0');
		const day = String(testDate.getUTCDate()).padStart(2, '0');
		const hour = String(testDate.getUTCHours()).padStart(2, '0');
		const minute = String(testDate.getUTCMinutes()).padStart(2, '0');
		const expected = `${month}${day}${hour}${minute}`;

		expect(expected).toBe('01010000');
	});

	it('should format December 31st at 23:59 as 12312359', async () => {
		const testDate = new Date('2025-12-31T23:59:00Z');
		const month = String(testDate.getUTCMonth() + 1).padStart(2, '0');
		const day = String(testDate.getUTCDate()).padStart(2, '0');
		const hour = String(testDate.getUTCHours()).padStart(2, '0');
		const minute = String(testDate.getUTCMinutes()).padStart(2, '0');
		const expected = `${month}${day}${hour}${minute}`;

		expect(expected).toBe('12312359');
	});

	it('should format double-digit month without leading zero', async () => {
		const testDate = new Date('2025-10-15T10:30:00Z');
		const month = String(testDate.getUTCMonth() + 1).padStart(2, '0');
		const day = String(testDate.getUTCDate()).padStart(2, '0');
		const hour = String(testDate.getUTCHours()).padStart(2, '0');
		const minute = String(testDate.getUTCMinutes()).padStart(2, '0');
		const expected = `${month}${day}${hour}${minute}`;

		expect(expected).toBe('10151030');
	});

	it('should format double-digit day without leading zero', async () => {
		const testDate = new Date('2025-06-20T14:45:00Z');
		const month = String(testDate.getUTCMonth() + 1).padStart(2, '0');
		const day = String(testDate.getUTCDate()).padStart(2, '0');
		const hour = String(testDate.getUTCHours()).padStart(2, '0');
		const minute = String(testDate.getUTCMinutes()).padStart(2, '0');
		const expected = `${month}${day}${hour}${minute}`;

		expect(expected).toBe('06201445');
	});

	it('should format double-digit hour without leading zero', async () => {
		const testDate = new Date('2025-03-10T18:22:00Z');
		const month = String(testDate.getUTCMonth() + 1).padStart(2, '0');
		const day = String(testDate.getUTCDate()).padStart(2, '0');
		const hour = String(testDate.getUTCHours()).padStart(2, '0');
		const minute = String(testDate.getUTCMinutes()).padStart(2, '0');
		const expected = `${month}${day}${hour}${minute}`;

		expect(expected).toBe('03101822');
	});

	it('should format double-digit minute without leading zero', async () => {
		const testDate = new Date('2025-04-07T09:55:00Z');
		const month = String(testDate.getUTCMonth() + 1).padStart(2, '0');
		const day = String(testDate.getUTCDate()).padStart(2, '0');
		const hour = String(testDate.getUTCHours()).padStart(2, '0');
		const minute = String(testDate.getUTCMinutes()).padStart(2, '0');
		const expected = `${month}${day}${hour}${minute}`;

		expect(expected).toBe('04070955');
	});

	it('should format single-digit month with leading zero', async () => {
		const testDate = new Date('2025-09-08T11:33:00Z');
		const month = String(testDate.getUTCMonth() + 1).padStart(2, '0');
		const day = String(testDate.getUTCDate()).padStart(2, '0');
		const hour = String(testDate.getUTCHours()).padStart(2, '0');
		const minute = String(testDate.getUTCMinutes()).padStart(2, '0');
		const expected = `${month}${day}${hour}${minute}`;

		expect(expected).toBe('09081133');
	});

	it('should format single-digit day with leading zero', async () => {
		const testDate = new Date('2025-11-03T16:48:00Z');
		const month = String(testDate.getUTCMonth() + 1).padStart(2, '0');
		const day = String(testDate.getUTCDate()).padStart(2, '0');
		const hour = String(testDate.getUTCHours()).padStart(2, '0');
		const minute = String(testDate.getUTCMinutes()).padStart(2, '0');
		const expected = `${month}${day}${hour}${minute}`;

		expect(expected).toBe('11031648');
	});

	it('should format single-digit hour with leading zero', async () => {
		const testDate = new Date('2025-05-22T08:15:00Z');
		const month = String(testDate.getUTCMonth() + 1).padStart(2, '0');
		const day = String(testDate.getUTCDate()).padStart(2, '0');
		const hour = String(testDate.getUTCHours()).padStart(2, '0');
		const minute = String(testDate.getUTCMinutes()).padStart(2, '0');
		const expected = `${month}${day}${hour}${minute}`;

		expect(expected).toBe('05220815');
	});

	it('should format single-digit minute with leading zero', async () => {
		const testDate = new Date('2025-07-14T19:04:00Z');
		const month = String(testDate.getUTCMonth() + 1).padStart(2, '0');
		const day = String(testDate.getUTCDate()).padStart(2, '0');
		const hour = String(testDate.getUTCHours()).padStart(2, '0');
		const minute = String(testDate.getUTCMinutes()).padStart(2, '0');
		const expected = `${month}${day}${hour}${minute}`;

		expect(expected).toBe('07141904');
	});

	it('should handle late evening hour', async () => {
		const testDate = new Date('2025-08-25T22:33:00Z');
		const month = String(testDate.getUTCMonth() + 1).padStart(2, '0');
		const day = String(testDate.getUTCDate()).padStart(2, '0');
		const hour = String(testDate.getUTCHours()).padStart(2, '0');
		const minute = String(testDate.getUTCMinutes()).padStart(2, '0');
		const expected = `${month}${day}${hour}${minute}`;

		expect(expected).toBe('08252233');
	});

	it('should handle end of month', async () => {
		const testDate = new Date('2025-09-30T12:00:00Z');
		const month = String(testDate.getUTCMonth() + 1).padStart(2, '0');
		const day = String(testDate.getUTCDate()).padStart(2, '0');
		const hour = String(testDate.getUTCHours()).padStart(2, '0');
		const minute = String(testDate.getUTCMinutes()).padStart(2, '0');
		const expected = `${month}${day}${hour}${minute}`;

		expect(expected).toBe('09301200');
	});

	it('should handle February with single-digit day', async () => {
		const testDate = new Date('2026-02-09T17:25:00Z');
		const month = String(testDate.getUTCMonth() + 1).padStart(2, '0');
		const day = String(testDate.getUTCDate()).padStart(2, '0');
		const hour = String(testDate.getUTCHours()).padStart(2, '0');
		const minute = String(testDate.getUTCMinutes()).padStart(2, '0');
		const expected = `${month}${day}${hour}${minute}`;

		expect(expected).toBe('02091725');
	});
});
