upgrade-go-deps:
    go get -u all

[working-directory('cmd/webapp')]
upgrade-svelte-deps:
    npm upgrade

upgrade-deps: upgrade-go-deps upgrade-svelte-deps
