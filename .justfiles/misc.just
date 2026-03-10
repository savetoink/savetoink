version:
    @echo "$(cat VERSION)-$(date -u +%m%d%H%M)-$(git rev-parse --short HEAD)"

[working-directory("tools")]
tools-init:
    @bun install

# Generate icons from https://www.svgrepo.com/svg/168917/ereader SVG source
generate-icons: tools-init generate-icons-webapp generate-icons-website generate-icons-extension

[working-directory("frontend/webapp")]
generate-icons-webapp:
    @bun run icons:generate

[working-directory("frontend/website")]
generate-icons-website:
    @bun run icons:generate

[working-directory("frontend/extension")]
generate-icons-extension:
    @bun run icons:generate

# Generate TypeScript types from OpenAPI specification
gen-api-types:
    @bunx openapi-typescript backend/lib/server/openapi.yaml -o frontend/shared/src/types/api-generated.ts

# Update VERSION file and all package.json files with new version
[working-directory("tools")]
update-versions version_type:
    @bun i
    @bun update_versions.ts {{ version_type }}

bump-version version_type:
    @just update-versions {{ version_type }}
    @git commit -m "chore: bump versions to $(cat VERSION)" .

install-hooks:
    @lefthook install

init-frontend: init-extension

dotenv ENV:
    @rm -f .env
    @ln -s .env.{{ ENV }} .env
