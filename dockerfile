FROM golang:1.19 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o cmd/main .

FROM redis:latest
COPY --from=builder /app/main .
EXPOSE 6379
EXPOSE 8080

CMD redis-server --daemonize yes && ./main