FROM golang:1.21.1-bullseye

WORKDIR /bot
COPY . .

RUN go mod download
RUN go build ./cmd/botsu

RUN wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/local/bin/yt-dlp
RUN chmod a+rx /usr/local/bin/yt-dlp

CMD ["./botsu"]
