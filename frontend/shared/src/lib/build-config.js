import { readFileSync } from 'fs';
import { execSync } from 'child_process';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export function getVersion() {
	return readFileSync(path.resolve(__dirname, '../../../../VERSION'), 'utf-8').trim();
}

export function getBuildDate() {
	return new Date().toISOString().slice(0, 10).replace(/-/g, '');
}

export function getGitHash() {
	return execSync('git rev-parse --short HEAD', {
		cwd: path.resolve(__dirname, '../../../../')
	})
		.toString()
		.trim();
}

export function getBuildConfig() {
	return {
		version: getVersion(),
		buildDate: getBuildDate(),
		gitHash: getGitHash()
	};
}
