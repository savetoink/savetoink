bootstrap-extension:
    bunx wxt@latest init
    cd frontend/extension && bun init

[working-directory('frontend/extension')]
server-extension:
    bun install
    bun run dev

[working-directory('frontend/extension')]
build-extension:
    bun install
    bun run build
    bun run build:firefox

[working-directory('frontend/extension')]
server-extension-firefox:
    bun install
    bun run dev:firefox
