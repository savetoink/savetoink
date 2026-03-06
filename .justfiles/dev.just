server-http:
    reflex -r '\.(env|go)$' -s -- go run -ldflags "{{ build_flags }}=$(just version)" backend/cmd/http/main.go

[working-directory('frontend/webapp')]
server-webapp:
    bun run dev --open

[working-directory('frontend/website')]
server-website:
    bun run dev
