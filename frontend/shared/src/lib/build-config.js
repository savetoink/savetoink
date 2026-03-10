import { readFileSync } from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export function getVersion() {
	return readFileSync(path.resolve(__dirname, '../../../../VERSION'), 'utf-8').trim();
}

export function getBuildConfig() {
	return {
		version: getVersion()
	};
}
