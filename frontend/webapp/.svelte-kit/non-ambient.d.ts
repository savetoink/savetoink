
// this file is generated — do not edit it


declare module "svelte/elements" {
	export interface HTMLAttributes<T> {
		'data-sveltekit-keepfocus'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-noscroll'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-preload-code'?:
			| true
			| ''
			| 'eager'
			| 'viewport'
			| 'hover'
			| 'tap'
			| 'off'
			| undefined
			| null;
		'data-sveltekit-preload-data'?: true | '' | 'hover' | 'tap' | 'off' | undefined | null;
		'data-sveltekit-reload'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-replacestate'?: true | '' | 'off' | undefined | null;
	}
}

export {};


declare module "$app/types" {
	type MatcherParam<M> = M extends (param : string) => param is (infer U extends string) ? U : string;

	export interface AppTypes {
		RouteId(): "/" | "/account" | "/articles" | "/articles/[id]" | "/auth" | "/auth/callback" | "/new" | "/sentry" | "/sentry/tunnel";
		RouteParams(): {
			"/articles/[id]": { id: string }
		};
		LayoutParams(): {
			"/": { id?: string };
			"/account": Record<string, never>;
			"/articles": { id?: string };
			"/articles/[id]": { id: string };
			"/auth": Record<string, never>;
			"/auth/callback": Record<string, never>;
			"/new": Record<string, never>;
			"/sentry": Record<string, never>;
			"/sentry/tunnel": Record<string, never>
		};
		Pathname(): "/" | "/account" | "/articles" | `/articles/${string}` & {} | "/auth/callback" | "/new" | "/sentry/tunnel";
		ResolvedPathname(): `${"" | `/${string}`}${ReturnType<AppTypes['Pathname']>}`;
		Asset(): "/android-icon-192x192.png" | "/android-icon-512x512.png" | "/apple-touch-icon-120x120.png" | "/apple-touch-icon-144x144.png" | "/apple-touch-icon-152x152.png" | "/apple-touch-icon-60x60.png" | "/apple-touch-icon-76x76.png" | "/apple-touch-icon.png" | "/favicon.ico" | "/favicon.svg" | "/manifest.json" | "/robots.txt" | string & {};
	}
}