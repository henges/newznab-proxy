FROM golang:alpine AS build

WORKDIR /app
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go generate ./...
RUN go build -o app .

FROM alpine
WORKDIR /app
COPY --from=build /app/app ./app

ENTRYPOINT ["/app/app"]
