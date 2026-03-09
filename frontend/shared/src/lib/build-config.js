import { readFileSync } from 'fs';
import { execSync } from 'child_process';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export function getVersion() {
	return readFileSync(path.resolve(__dirname, '../../../../VERSION'), 'utf-8').trim();
}

export function getBuildDate() {
	const now = new Date();
	const month = String(now.getUTCMonth() + 1).padStart(2, '0');
	const day = String(now.getUTCDate()).padStart(2, '0');
	const hour = String(now.getUTCHours()).padStart(2, '0');
	const minute = String(now.getUTCMinutes()).padStart(2, '0');
	return `${month}${day}${hour}${minute}`;
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
