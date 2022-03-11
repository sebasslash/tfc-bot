FROM golang:1.17.5-buster AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
ENV CGO_ENABLED=0
RUN go build -o /tfc-bot

FROM alpine:3.15

WORKDIR /

COPY --from=build /tfc-bot /tfc-bot

EXPOSE 10000

CMD ["/tfc-bot"]
