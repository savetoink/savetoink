[working-directory('frontend/extension')]
bootstrap-extension:
    bunx wxt@latest init

[working-directory('frontend/extension')]
init-extension:
    bun init

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
