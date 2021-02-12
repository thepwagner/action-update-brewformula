FROM golang:1.15.8 AS builder

WORKDIR /app
COPY go.mod /app
COPY go.sum /app
RUN go mod download

COPY . /app
RUN go build -o /update-brewformula /app

FROM debian:buster-slim
COPY --from=builder /update-brewformula /update-brewformula
ENTRYPOINT ["/update-brewformula"]
