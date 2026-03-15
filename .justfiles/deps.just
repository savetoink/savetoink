upgrade-go-deps:
    go get -u all

[working-directory('frontend/shared')]
upgrade-shared-deps:
    @bun outdated
    @bun update --latest --filter '!@savetoink/shared'

[working-directory('frontend/webapp')]
upgrade-webapp-deps:
    @bun outdated
    @bun update --latest --filter '!@savetoink/shared'

[working-directory('frontend/website')]
upgrade-website-deps:
    @bun outdated
    @bun update --latest --filter '!@savetoink/shared'

[working-directory('frontend/extension')]
upgrade-extension-deps:
    @bun outdated
    @bun update --latest --filter '!@savetoink/shared'

upgrade-frontend-deps: upgrade-webapp-deps upgrade-shared-deps upgrade-website-deps upgrade-extension-deps

upgrade-deps: upgrade-go-deps upgrade-frontend-deps
