bootstrap-extension:
    bunx wxt@latest init
    cd cmd/extension && bun init

[working-directory('cmd/extension')]
server-extension:
    bun run dev
