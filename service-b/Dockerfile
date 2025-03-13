FROM golang:1.24-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o service-b .

# Run tests with mock services
#ENV TEST_MODE=true
#RUN go test -v ./...

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/service-b .

EXPOSE 8081
CMD ["./service-b"]
