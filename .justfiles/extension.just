bootstrap-extension:
    bunx wxt@latest init
    cd cmd/extension && bun init

[working-directory('cmd/extension')]
server-extension:
    bun install
    bun run dev

[working-directory('cmd/extension')]
build-extension:
    bun install
    bun run build
    bun run build:firefox

[working-directory('cmd/extension')]
server-extension-firefox:
    bun install
    bun run dev:firefox
