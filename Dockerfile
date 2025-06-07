FROM golang:1.24-alpine AS builder
WORKDIR /app
RUN go install github.com/swaggo/swag/cmd/swag@latest
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN swag init -g main.go -o ./docs
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o goauth ./

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/goauth .
COPY --from=builder /app/docs ./docs
CMD ["./goauth"]
