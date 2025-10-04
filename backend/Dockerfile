# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN mkdir -p data
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server

FROM gcr.io/distroless/static:nonroot
WORKDIR /
ENV PORT=8080
ENV SQLITE_PATH=/data/app.sqlite
COPY --from=build /out/server /server
COPY --from=build /app/data /data
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
