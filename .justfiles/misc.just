version:
    @echo "$(cat VERSION)-$(date -u +%Y%m%d)-$(git rev-parse --short HEAD)"
