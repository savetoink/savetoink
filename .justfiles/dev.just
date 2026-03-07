[working-directory('backend')]
server-http:
    reflex -r '\.go$' -s -- go run -ldflags "{{ build_flags }}=$(just version)" cmd/http/main.go

[working-directory('frontend/webapp')]
server-webapp:
    bun run dev --open

[working-directory('frontend/website')]
server-website:
    bun run dev
