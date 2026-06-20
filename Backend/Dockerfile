FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o autorent-backend .

FROM alpine:3.24

RUN apk add --no-cache ca-certificates wget

WORKDIR /app

COPY --from=builder /app/autorent-backend ./autorent-backend

EXPOSE 8080

CMD ["./autorent-backend"]
