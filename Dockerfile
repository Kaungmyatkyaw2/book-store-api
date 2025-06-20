FROM golang:1.24.3-alpine

WORKDIR /app 

COPY go.mod ./
RUN go mod download

COPY . . 

RUN go build -o main ./cmd/api

EXPOSE 4000

CMD ["/app/main"]
