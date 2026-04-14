import { defineConfig, loadEnv } from "vite";
import { URL } from "node:url";

export default defineConfig(({ mode }) => {
	const env = loadEnv(mode, process.cwd(), "");

	let serverConfig: { host?: string; port?: number } = {};

	try {
		const url = new URL(env.PUBLIC_APP_URL || "");
		serverConfig = {
			host: url.hostname,
			port: Number.parseInt(url.port, 10) || 5173,
		};
	} catch {
		serverConfig = { port: 5173 };
	}

	return {
		base: "./",
		server: serverConfig,
		build: {
			target: "esnext",
			outDir: "dist",
			emptyOutDir: true,
		},
		define: {
			"import.meta.env.PUBLIC_APP_URL": JSON.stringify(env.PUBLIC_APP_URL || ""),
		},
	};
});
