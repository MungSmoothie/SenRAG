FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o senrag .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/senrag .
COPY --from=builder /app/config.yaml .

RUN mkdir -p /app/uploads

EXPOSE 8080

CMD ["./senrag"]
