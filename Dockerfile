FROM golang:1.25.1-alpine as Builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/subscriptions-service ./cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/bin/subscriptions-service .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./subscriptions-service"]
