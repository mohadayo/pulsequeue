FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod .
COPY *.go .
RUN go build -o worker-engine .

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/worker-engine .

EXPOSE 8081

CMD ["./worker-engine"]
