bootstrap-extension:
    bunx wxt@latest init
    cd cmd/web/extension && bun init

[working-directory('cmd/web/extension')]
server-extension:
    bun install
    bun run dev

[working-directory('cmd/web/extension')]
build-extension:
    bun install
    bun run build
    bun run build:firefox

[working-directory('cmd/web/extension')]
server-extension-firefox:
    bun install
    bun run dev:firefox
