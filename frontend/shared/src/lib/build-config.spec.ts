import { describe, expect, it } from 'vitest';

describe('getVersion', () => {
	it('should return the version from VERSION file', async () => {
		const { getVersion } = await import('./build-config.js');
		const result = getVersion();

		expect(result).toBeTypeOf('string');
		expect(result).toMatch(/^\d+\.\d+\.\d+(-dev\.\d{8}\.[a-f0-9]{7})?$/);
	});

	it('should not be empty', async () => {
		const { getVersion } = await import('./build-config.js');
		const result = getVersion();

		expect(result.length).toBeGreaterThan(0);
	});
});

describe('getBuildConfig', () => {
	it('should return an object with version property', async () => {
		const { getBuildConfig } = await import('./build-config.js');
		const result = getBuildConfig();

		expect(result).toHaveProperty('version');
		expect(result.version).toBeTypeOf('string');
	});

	it('should have version that matches getVersion', async () => {
		const { getBuildConfig, getVersion } = await import('./build-config.js');
		const config = getBuildConfig();
		const version = getVersion();

		expect(config.version).toBe(version);
	});
});
