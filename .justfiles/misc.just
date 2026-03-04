version:
    @echo "$(cat VERSION)-$(date -u +%Y%m%d)-$(git rev-parse --short HEAD)"

[working-directory("tools")]
tools-init:
    @bun install

# Generate icons from https://www.svgrepo.com/svg/168917/ereader SVG source
generate-icons: tools-init generate-icons-webapp generate-icons-website generate-icons-extension

[working-directory("cmd/web/webapp")]
generate-icons-webapp:
    @bun run icons:generate

[working-directory("cmd/web/website")]
generate-icons-website:
    @bun run icons:generate

[working-directory("cmd/web/extension")]
generate-icons-extension:
    @bun run icons:generate

# Update VERSION file and all package.json files with new version
[working-directory("tools")]
bump-versions version_type:
    @bun run update_versions.ts {{ version_type }}
