bootstrap-extension:
    bunx wxt@latest init
    cd cmd/extension && bun init

[working-directory('cmd/extension')]
server-extension:
    bun run dev

[working-directory('cmd/extension')]
build-extension:
    bun run build
    bun run build:firefox

[working-directory('cmd/extension')]
server-extension-firefox:
    bun run dev:firefox
