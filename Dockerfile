FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /couchcade ./cmd/couchcade

FROM scratch
COPY --from=build /couchcade /couchcade
EXPOSE 8080
ENTRYPOINT ["/couchcade", "serve", "-addr", ":8080"]
