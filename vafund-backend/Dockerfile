FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o vafund-backend main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/vafund-backend ./vafund-backend
COPY --from=builder /app/.env ./
COPY --from=builder /app/crypto ./crypto
EXPOSE 3000
CMD ["./vafund-backend"]
