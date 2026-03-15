[working-directory('backend')]
http:
    reflex -r '\.go$' -s -- go run cmd/http/main.go

[working-directory('frontend/webapp')]
webapp:
    bun run dev --open

[working-directory('frontend/website')]
website:
    bun run dev --open
