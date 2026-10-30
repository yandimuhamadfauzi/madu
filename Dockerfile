FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/madu ./cmd/server

FROM alpine:3
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/madu /app/madu
COPY web /app/web
USER app
ENV GIN_MODE=release
EXPOSE 8080
CMD ["/app/madu"]
