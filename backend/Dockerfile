FROM golang:latest AS build
ARG APP
RUN apt update && apt install -y just
COPY VERSION /app/
COPY go.sum go.mod /app/
COPY justfile /app/
COPY .justfiles /app/.justfiles
COPY cli /app/cli
COPY backend /app/backend
WORKDIR /app
RUN just build-$APP

FROM scratch AS runtime
COPY --from=build /etc/passwd /etc/passwd
# thanks to https://blog.wollomatic.de/posts/2025-01-28-go-tls-certificates/
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /app/bin/savetoink /app/bin/savetoink
USER nobody
WORKDIR /app/data
VOLUME ["/app/data"]
ENTRYPOINT ["/app/bin/savetoink"]
