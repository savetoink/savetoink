
// this file is generated — do not edit it


/// <reference types="@sveltejs/kit" />

/**
 * This module provides access to environment variables that are injected _statically_ into your bundle at build time and are limited to _private_ access.
 * 
 * |         | Runtime                                                                    | Build time                                                               |
 * | ------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
 * | Private | [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private) | [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private) |
 * | Public  | [`$env/dynamic/public`](https://svelte.dev/docs/kit/$env-dynamic-public)   | [`$env/static/public`](https://svelte.dev/docs/kit/$env-static-public)   |
 * 
 * Static environment variables are [loaded by Vite](https://vitejs.dev/guide/env-and-mode.html#env-files) from `.env` files and `process.env` at build time and then statically injected into your bundle at build time, enabling optimisations like dead code elimination.
 * 
 * **_Private_ access:**
 * 
 * - This module cannot be imported into client-side code
 * - This module only includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://svelte.dev/docs/kit/configuration#env) (if configured)
 * 
 * For example, given the following build time environment:
 * 
 * ```env
 * ENVIRONMENT=production
 * PUBLIC_BASE_URL=http://site.com
 * ```
 * 
 * With the default `publicPrefix` and `privatePrefix`:
 * 
 * ```ts
 * import { ENVIRONMENT, PUBLIC_BASE_URL } from '$env/static/private';
 * 
 * console.log(ENVIRONMENT); // => "production"
 * console.log(PUBLIC_BASE_URL); // => throws error during build
 * ```
 * 
 * The above values will be the same _even if_ different values for `ENVIRONMENT` or `PUBLIC_BASE_URL` are set at runtime, as they are statically replaced in your code with their build time values.
 */
declare module '$env/static/private' {
	export const SHELL: string;
	export const SELENIUM_JAR_PATH: string;
	export const CONDA: string;
	export const GITHUB_WORKSPACE: string;
	export const JAVA_HOME_11_X64: string;
	export const JOBS: string;
	export const JAVA_HOME_25_X64: string;
	export const INPUT_TRIGGER: string;
	export const INPUT_MODEL: string;
	export const NODE: string;
	export const GITHUB_PATH: string;
	export const GITHUB_ACTION: string;
	export const JAVA_HOME: string;
	export const GITHUB_RUN_NUMBER: string;
	export const RUNNER_NAME: string;
	export const GRADLE_HOME: string;
	export const GITHUB_REPOSITORY_OWNER_ID: string;
	export const ACTIONS_RUNNER_ACTION_ARCHIVE_CACHE: string;
	export const XDG_CONFIG_HOME: string;
	export const MEMORY_PRESSURE_WRITE: string;
	export const DOTNET_SKIP_FIRST_TIME_EXPERIENCE: string;
	export const ANT_HOME: string;
	export const npm_config_local_prefix: string;
	export const JAVA_HOME_8_X64: string;
	export const GITHUB_TRIGGERING_ACTOR: string;
	export const GITHUB_REF_TYPE: string;
	export const ACTIONS_CACHE_URL: string;
	export const HOMEBREW_CLEANUP_PERIODIC_FULL_DAYS: string;
	export const ANDROID_NDK: string;
	export const BOOTSTRAP_HASKELL_NONINTERACTIVE: string;
	export const PIPX_BIN_DIR: string;
	export const PWD: string;
	export const LOGNAME: string;
	export const GITHUB_REPOSITORY_ID: string;
	export const GITHUB_ACTIONS: string;
	export const USE_BAZEL_FALLBACK_VERSION: string;
	export const ANDROID_NDK_LATEST_HOME: string;
	export const BUN_WHICH_IGNORE_CWD: string;
	export const SYSTEMD_EXEC_PID: string;
	export const GITHUB_SHA: string;
	export const GITHUB_WORKFLOW_REF: string;
	export const POWERSHELL_DISTRIBUTION_CHANNEL: string;
	export const _: string;
	export const RUNNER_ENVIRONMENT: string;
	export const DOTNET_MULTILEVEL_LOOKUP: string;
	export const GITHUB_REF: string;
	export const RUNNER_OS: string;
	export const GITHUB_REF_PROTECTED: string;
	export const HOME: string;
	export const GITHUB_API_URL: string;
	export const LANG: string;
	export const GOROOT_1_25_X64: string;
	export const RUNNER_TRACKING_ID: string;
	export const RUNNER_ARCH: string;
	export const MEMORY_PRESSURE_WATCH: string;
	export const ACTIONS_CACHE_SERVICE_V2: string;
	export const RUNNER_TEMP: string;
	export const ACTIONS_RUNTIME_URL: string;
	export const GITHUB_STATE: string;
	export const EDGEWEBDRIVER: string;
	export const JAVA_HOME_21_X64: string;
	export const GITHUB_ENV: string;
	export const INVOCATION_ID: string;
	export const GITHUB_EVENT_PATH: string;
	export const GITHUB_EVENT_NAME: string;
	export const GITHUB_RUN_ID: string;
	export const JAVA_HOME_17_X64: string;
	export const INPUT_TOKEN: string;
	export const INIT_CWD: string;
	export const ANDROID_NDK_HOME: string;
	export const GITHUB_STEP_SUMMARY: string;
	export const HOMEBREW_NO_AUTO_UPDATE: string;
	export const GITHUB_ACTOR: string;
	export const NVM_DIR: string;
	export const SGX_AESM_ADDR: string;
	export const GITHUB_RUN_ATTEMPT: string;
	export const ACTIONS_RESULTS_URL: string;
	export const ANDROID_HOME: string;
	export const GITHUB_GRAPHQL_URL: string;
	export const npm_package_name: string;
	export const ACCEPT_EULA: string;
	export const ACTIONS_RUNTIME_TOKEN: string;
	export const USER: string;
	export const INPUT_GITHUB_TOKEN: string;
	export const PSModulePath: string;
	export const GITHUB_SERVER_URL: string;
	export const PIPX_HOME: string;
	export const GECKOWEBDRIVER: string;
	export const CHROMEWEBDRIVER: string;
	export const SHLVL: string;
	export const ANDROID_SDK_ROOT: string;
	export const VCPKG_INSTALLATION_ROOT: string;
	export const GITHUB_ACTOR_ID: string;
	export const ACTIONS_ORCHESTRATION_ID: string;
	export const RUNNER_TOOL_CACHE: string;
	export const ImageVersion: string;
	export const DOTNET_NOLOGO: string;
	export const GOROOT_1_23_X64: string;
	export const GITHUB_WORKFLOW_SHA: string;
	export const GOROOT_1_24_X64: string;
	export const GITHUB_REF_NAME: string;
	export const npm_config_user_agent: string;
	export const npm_execpath: string;
	export const GITHUB_JOB: string;
	export const XDG_RUNTIME_DIR: string;
	export const AZURE_EXTENSION_DIR: string;
	export const npm_package_json: string;
	export const GITHUB_REPOSITORY: string;
	export const GOROOT_1_22_X64: string;
	export const ANDROID_NDK_ROOT: string;
	export const CHROME_BIN: string;
	export const GITHUB_RETENTION_DAYS: string;
	export const JOURNAL_STREAM: string;
	export const RUNNER_WORKSPACE: string;
	export const GITHUB_ACTION_REPOSITORY: string;
	export const INPUT_PROVIDER: string;
	export const PATH: string;
	export const GITHUB_BASE_REF: string;
	export const GHCUP_INSTALL_BASE_PREFIX: string;
	export const CI: string;
	export const SWIFT_PATH: string;
	export const ImageOS: string;
	export const GITHUB_REPOSITORY_OWNER: string;
	export const GITHUB_HEAD_REF: string;
	export const GITHUB_ACTION_REF: string;
	export const ENABLE_RUNNER_TRACING: string;
	export const INPUT_PROMPT: string;
	export const GITHUB_WORKFLOW: string;
	export const DEBIAN_FRONTEND: string;
	export const INPUT_THINKING_LEVEL: string;
	export const npm_node_execpath: string;
	export const GITHUB_OUTPUT: string;
	export const AGENT_TOOLSDIRECTORY: string;
	export const OLDPWD: string;
}

/**
 * This module provides access to environment variables that are injected _statically_ into your bundle at build time and are _publicly_ accessible.
 * 
 * |         | Runtime                                                                    | Build time                                                               |
 * | ------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
 * | Private | [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private) | [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private) |
 * | Public  | [`$env/dynamic/public`](https://svelte.dev/docs/kit/$env-dynamic-public)   | [`$env/static/public`](https://svelte.dev/docs/kit/$env-static-public)   |
 * 
 * Static environment variables are [loaded by Vite](https://vitejs.dev/guide/env-and-mode.html#env-files) from `.env` files and `process.env` at build time and then statically injected into your bundle at build time, enabling optimisations like dead code elimination.
 * 
 * **_Public_ access:**
 * 
 * - This module _can_ be imported into client-side code
 * - **Only** variables that begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) (which defaults to `PUBLIC_`) are included
 * 
 * For example, given the following build time environment:
 * 
 * ```env
 * ENVIRONMENT=production
 * PUBLIC_BASE_URL=http://site.com
 * ```
 * 
 * With the default `publicPrefix` and `privatePrefix`:
 * 
 * ```ts
 * import { ENVIRONMENT, PUBLIC_BASE_URL } from '$env/static/public';
 * 
 * console.log(ENVIRONMENT); // => throws error during build
 * console.log(PUBLIC_BASE_URL); // => "http://site.com"
 * ```
 * 
 * The above values will be the same _even if_ different values for `ENVIRONMENT` or `PUBLIC_BASE_URL` are set at runtime, as they are statically replaced in your code with their build time values.
 */
declare module '$env/static/public' {
	
}

/**
 * This module provides access to environment variables set _dynamically_ at runtime and that are limited to _private_ access.
 * 
 * |         | Runtime                                                                    | Build time                                                               |
 * | ------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
 * | Private | [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private) | [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private) |
 * | Public  | [`$env/dynamic/public`](https://svelte.dev/docs/kit/$env-dynamic-public)   | [`$env/static/public`](https://svelte.dev/docs/kit/$env-static-public)   |
 * 
 * Dynamic environment variables are defined by the platform you're running on. For example if you're using [`adapter-node`](https://github.com/sveltejs/kit/tree/main/packages/adapter-node) (or running [`vite preview`](https://svelte.dev/docs/kit/cli)), this is equivalent to `process.env`.
 * 
 * **_Private_ access:**
 * 
 * - This module cannot be imported into client-side code
 * - This module includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://svelte.dev/docs/kit/configuration#env) (if configured)
 * 
 * > [!NOTE] In `dev`, `$env/dynamic` includes environment variables from `.env`. In `prod`, this behavior will depend on your adapter.
 * 
 * > [!NOTE] To get correct types, environment variables referenced in your code should be declared (for example in an `.env` file), even if they don't have a value until the app is deployed:
 * >
 * > ```env
 * > MY_FEATURE_FLAG=
 * > ```
 * >
 * > You can override `.env` values from the command line like so:
 * >
 * > ```sh
 * > MY_FEATURE_FLAG="enabled" npm run dev
 * > ```
 * 
 * For example, given the following runtime environment:
 * 
 * ```env
 * ENVIRONMENT=production
 * PUBLIC_BASE_URL=http://site.com
 * ```
 * 
 * With the default `publicPrefix` and `privatePrefix`:
 * 
 * ```ts
 * import { env } from '$env/dynamic/private';
 * 
 * console.log(env.ENVIRONMENT); // => "production"
 * console.log(env.PUBLIC_BASE_URL); // => undefined
 * ```
 */
declare module '$env/dynamic/private' {
	export const env: {
		SHELL: string;
		SELENIUM_JAR_PATH: string;
		CONDA: string;
		GITHUB_WORKSPACE: string;
		JAVA_HOME_11_X64: string;
		JOBS: string;
		JAVA_HOME_25_X64: string;
		INPUT_TRIGGER: string;
		INPUT_MODEL: string;
		NODE: string;
		GITHUB_PATH: string;
		GITHUB_ACTION: string;
		JAVA_HOME: string;
		GITHUB_RUN_NUMBER: string;
		RUNNER_NAME: string;
		GRADLE_HOME: string;
		GITHUB_REPOSITORY_OWNER_ID: string;
		ACTIONS_RUNNER_ACTION_ARCHIVE_CACHE: string;
		XDG_CONFIG_HOME: string;
		MEMORY_PRESSURE_WRITE: string;
		DOTNET_SKIP_FIRST_TIME_EXPERIENCE: string;
		ANT_HOME: string;
		npm_config_local_prefix: string;
		JAVA_HOME_8_X64: string;
		GITHUB_TRIGGERING_ACTOR: string;
		GITHUB_REF_TYPE: string;
		ACTIONS_CACHE_URL: string;
		HOMEBREW_CLEANUP_PERIODIC_FULL_DAYS: string;
		ANDROID_NDK: string;
		BOOTSTRAP_HASKELL_NONINTERACTIVE: string;
		PIPX_BIN_DIR: string;
		PWD: string;
		LOGNAME: string;
		GITHUB_REPOSITORY_ID: string;
		GITHUB_ACTIONS: string;
		USE_BAZEL_FALLBACK_VERSION: string;
		ANDROID_NDK_LATEST_HOME: string;
		BUN_WHICH_IGNORE_CWD: string;
		SYSTEMD_EXEC_PID: string;
		GITHUB_SHA: string;
		GITHUB_WORKFLOW_REF: string;
		POWERSHELL_DISTRIBUTION_CHANNEL: string;
		_: string;
		RUNNER_ENVIRONMENT: string;
		DOTNET_MULTILEVEL_LOOKUP: string;
		GITHUB_REF: string;
		RUNNER_OS: string;
		GITHUB_REF_PROTECTED: string;
		HOME: string;
		GITHUB_API_URL: string;
		LANG: string;
		GOROOT_1_25_X64: string;
		RUNNER_TRACKING_ID: string;
		RUNNER_ARCH: string;
		MEMORY_PRESSURE_WATCH: string;
		ACTIONS_CACHE_SERVICE_V2: string;
		RUNNER_TEMP: string;
		ACTIONS_RUNTIME_URL: string;
		GITHUB_STATE: string;
		EDGEWEBDRIVER: string;
		JAVA_HOME_21_X64: string;
		GITHUB_ENV: string;
		INVOCATION_ID: string;
		GITHUB_EVENT_PATH: string;
		GITHUB_EVENT_NAME: string;
		GITHUB_RUN_ID: string;
		JAVA_HOME_17_X64: string;
		INPUT_TOKEN: string;
		INIT_CWD: string;
		ANDROID_NDK_HOME: string;
		GITHUB_STEP_SUMMARY: string;
		HOMEBREW_NO_AUTO_UPDATE: string;
		GITHUB_ACTOR: string;
		NVM_DIR: string;
		SGX_AESM_ADDR: string;
		GITHUB_RUN_ATTEMPT: string;
		ACTIONS_RESULTS_URL: string;
		ANDROID_HOME: string;
		GITHUB_GRAPHQL_URL: string;
		npm_package_name: string;
		ACCEPT_EULA: string;
		ACTIONS_RUNTIME_TOKEN: string;
		USER: string;
		INPUT_GITHUB_TOKEN: string;
		PSModulePath: string;
		GITHUB_SERVER_URL: string;
		PIPX_HOME: string;
		GECKOWEBDRIVER: string;
		CHROMEWEBDRIVER: string;
		SHLVL: string;
		ANDROID_SDK_ROOT: string;
		VCPKG_INSTALLATION_ROOT: string;
		GITHUB_ACTOR_ID: string;
		ACTIONS_ORCHESTRATION_ID: string;
		RUNNER_TOOL_CACHE: string;
		ImageVersion: string;
		DOTNET_NOLOGO: string;
		GOROOT_1_23_X64: string;
		GITHUB_WORKFLOW_SHA: string;
		GOROOT_1_24_X64: string;
		GITHUB_REF_NAME: string;
		npm_config_user_agent: string;
		npm_execpath: string;
		GITHUB_JOB: string;
		XDG_RUNTIME_DIR: string;
		AZURE_EXTENSION_DIR: string;
		npm_package_json: string;
		GITHUB_REPOSITORY: string;
		GOROOT_1_22_X64: string;
		ANDROID_NDK_ROOT: string;
		CHROME_BIN: string;
		GITHUB_RETENTION_DAYS: string;
		JOURNAL_STREAM: string;
		RUNNER_WORKSPACE: string;
		GITHUB_ACTION_REPOSITORY: string;
		INPUT_PROVIDER: string;
		PATH: string;
		GITHUB_BASE_REF: string;
		GHCUP_INSTALL_BASE_PREFIX: string;
		CI: string;
		SWIFT_PATH: string;
		ImageOS: string;
		GITHUB_REPOSITORY_OWNER: string;
		GITHUB_HEAD_REF: string;
		GITHUB_ACTION_REF: string;
		ENABLE_RUNNER_TRACING: string;
		INPUT_PROMPT: string;
		GITHUB_WORKFLOW: string;
		DEBIAN_FRONTEND: string;
		INPUT_THINKING_LEVEL: string;
		npm_node_execpath: string;
		GITHUB_OUTPUT: string;
		AGENT_TOOLSDIRECTORY: string;
		OLDPWD: string;
		[key: `PUBLIC_${string}`]: undefined;
		[key: `${string}`]: string | undefined;
	}
}

/**
 * This module provides access to environment variables set _dynamically_ at runtime and that are _publicly_ accessible.
 * 
 * |         | Runtime                                                                    | Build time                                                               |
 * | ------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
 * | Private | [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private) | [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private) |
 * | Public  | [`$env/dynamic/public`](https://svelte.dev/docs/kit/$env-dynamic-public)   | [`$env/static/public`](https://svelte.dev/docs/kit/$env-static-public)   |
 * 
 * Dynamic environment variables are defined by the platform you're running on. For example if you're using [`adapter-node`](https://github.com/sveltejs/kit/tree/main/packages/adapter-node) (or running [`vite preview`](https://svelte.dev/docs/kit/cli)), this is equivalent to `process.env`.
 * 
 * **_Public_ access:**
 * 
 * - This module _can_ be imported into client-side code
 * - **Only** variables that begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) (which defaults to `PUBLIC_`) are included
 * 
 * > [!NOTE] In `dev`, `$env/dynamic` includes environment variables from `.env`. In `prod`, this behavior will depend on your adapter.
 * 
 * > [!NOTE] To get correct types, environment variables referenced in your code should be declared (for example in an `.env` file), even if they don't have a value until the app is deployed:
 * >
 * > ```env
 * > MY_FEATURE_FLAG=
 * > ```
 * >
 * > You can override `.env` values from the command line like so:
 * >
 * > ```sh
 * > MY_FEATURE_FLAG="enabled" npm run dev
 * > ```
 * 
 * For example, given the following runtime environment:
 * 
 * ```env
 * ENVIRONMENT=production
 * PUBLIC_BASE_URL=http://example.com
 * ```
 * 
 * With the default `publicPrefix` and `privatePrefix`:
 * 
 * ```ts
 * import { env } from '$env/dynamic/public';
 * console.log(env.ENVIRONMENT); // => undefined, not public
 * console.log(env.PUBLIC_BASE_URL); // => "http://example.com"
 * ```
 * 
 * ```
 * 
 * ```
 */
declare module '$env/dynamic/public' {
	export const env: {
		[key: `PUBLIC_${string}`]: string | undefined;
	}
}
