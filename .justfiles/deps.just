upgrade-go-deps:
    go get -u all

[working-directory('frontend/shared')]
upgrade-shared-deps:
    @bun update

[working-directory('frontend/webapp')]
upgrade-webapp-deps:
    @bun update

[working-directory('frontend/website')]
upgrade-website-deps:
    @bun update

[working-directory('frontend/extension')]
upgrade-extension-deps:
    @bun update

upgrade-deps: upgrade-go-deps upgrade-webapp-deps upgrade-shared-deps upgrade-website-deps upgrade-extension-deps
