FROM golang:1.26.4

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /neRugaysyaBot

COPY badWords.txt ./

CMD ["/neRugaysyaBot"]
