FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod ./
COPY . .

# Run tests with mock services
ENV TEST_MODE=true
RUN go test -v ./...

RUN go build -o main .

EXPOSE 8080

CMD ["./main"]