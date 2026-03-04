server-http:
    reflex -r '\.(env|go)$' -s -- go run -ldflags "{{ build_flags }}=$(just version)" ./cmd/http/main.go

[working-directory('cmd/web/webapp')]
server-webapp:
    bun run dev --open

[working-directory('cmd/web/website')]
server-website:
    npm run dev
