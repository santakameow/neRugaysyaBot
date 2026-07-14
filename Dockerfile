# stage 1
FROM golang:1.26.4 AS build
WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /neRugaysyaBot

# stage 2
FROM scratch
WORKDIR /app

COPY --from=build /neRugaysyaBot /neRugaysyaBot
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY badWords.txt ./

CMD ["/neRugaysyaBot"]
