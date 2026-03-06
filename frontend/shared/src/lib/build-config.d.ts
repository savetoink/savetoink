export function getVersion(): string;
export function getBuildDate(): string;
export function getGitHash(): string;

export interface BuildConfig {
  version: string;
  buildDate: string;
  gitHash: string;
}

export function getBuildConfig(): BuildConfig;
