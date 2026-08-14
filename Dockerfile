# stage 1
FROM golang:1.26.5 AS build
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o neRugaysyaBot

# stage 2
FROM gcr.io/distroless/static-debian12
WORKDIR /app

COPY --from=build /app/neRugaysyaBot ./

CMD ["./neRugaysyaBot"]
