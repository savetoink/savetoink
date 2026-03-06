[working-directory("frontend/website")]
install-webapp-deps:
    bun install

[working-directory("frontend/website")]
install-website-deps:
    bun install

[working-directory("frontend/extension")]
install-extension-deps:
    bun install

[working-directory("frontend/shared")]
install-frontend-deps: install-webapp-deps install-website-deps install-extension-deps
    bun install
