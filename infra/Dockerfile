FROM golang:latest AS build
ARG APP
RUN apt update && apt install -y just
COPY . /app
WORKDIR /app
RUN just build-$APP

FROM scratch AS runtime
# thanks to https://blog.wollomatic.de/posts/2025-01-28-go-tls-certificates/
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /app/bin/savetoink /app/bin/savetoink
ENTRYPOINT ["/app/bin/savetoink"]
