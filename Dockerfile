FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o kaskade

FROM alpine:latest AS runtime

WORKDIR /app

COPY --from=builder /app/kaskade /app/kaskade

EXPOSE 4000

CMD ["/app/kaskade"]
