FROM golang:1.23.4

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./todo-list

CMD ["./todo-list"] 