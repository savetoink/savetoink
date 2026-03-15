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
    @bunx openapi-typescript backend/lib/server/handlers/openapi.yaml -o frontend/shared/src/types/api-generated.ts

# Update VERSION file and all package.json files with new version
[working-directory("tools")]
update-versions version_type:
    @bun i
    @bun update_versions.ts {{ version_type }}

# Bump semver in VERSION and package.json files, create bump commit. Tag the commit with the new version if version_type is not "dev"
bump-version version_type="dev":
    @just update-versions {{ version_type }}
    @git commit -m "chore: bump versions to $(cat VERSION)" .
    @if [[ {{ version_type }} != "dev" ]]; then git tag "v$(cat VERSION)"; fi

bump-version-prod version_type="patch": check
    @just bump-version {{ version_type }}

install-hooks:
    @lefthook install

init-frontend: init-extension

dotenv ENV="local":
    @rm -f .env
    @ln -s .env.{{ ENV }} .env

zai-remaining:
    @bun run ./tools/zai_usage.ts
