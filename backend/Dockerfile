FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /out/backend ./cmd/main

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /out/backend /app/backend
COPY --from=builder /app/python_service /app/python_service

EXPOSE 4000

CMD ["/app/backend"]