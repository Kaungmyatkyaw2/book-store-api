FROM golang:1.22-alpine 

ENV GO111MODULE=on

WORKDIR /app 

COPY go.mod ./
RUN go mod download

COPY . . 

RUN go build -o main .

EXPOSE 4000

CMD ["./main"]
