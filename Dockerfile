FROM golang:alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o app .

FROM alpine
WORKDIR /app
COPY --from=build /app/app ./app

ENTRYPOINT ["/app/app"]
