FROM golang:1.21.1

WORKDIR /bot
COPY . .

RUN go mod download
RUN go build ./cmd/botsu

CMD ["./botsu"]
