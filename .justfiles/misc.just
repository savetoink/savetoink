version:
    @echo "$(cat VERSION)-$(date -u +%Y%m%d)-$(git rev-parse --short HEAD)"

# Generate icons from https://www.svgrepo.com/svg/168917/ereader SVG source
generate-icons:
    bun tools/generate_icons.ts
