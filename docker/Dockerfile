FROM golang:1.27.2-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go run github.com/swaggo/swag/cmd/swag@v1.16.2 init \
    -g main.go \
    -d ./cmd/api,./internal/httpapi \
    -o ./docs \
    --parseInternal

RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /out/api ./api

EXPOSE 8000

ENTRYPOINT ["./api"]
