FROM golang:1.21.1-bullseye

WORKDIR /bot

# Install dependencies first to make use of Docker cache
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o ./bin/botsu ./cmd/botsu

RUN wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/local/bin/yt-dlp
RUN chmod a+rx /usr/local/bin/yt-dlp

CMD ["./bin/botsu"]
