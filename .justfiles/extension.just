[working-directory('frontend/extension')]
bootstrap-extension:
    bunx wxt@latest init

[working-directory('frontend/extension')]
init-extension:
    bun init

[working-directory('frontend/extension')]
extension:
    bun install
    bun run dev

[working-directory('frontend/extension')]
extension-firefox:
    bun install
    bun run dev:firefox

[working-directory('frontend/extension')]
build-extension:
    bun install
    bun run build
    bun run build:firefox
    bun run zip
    bun run zip:firefox
