FROM golang:1.24-alpine AS base

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o expenses

EXPOSE 7074

CMD ["/build/expenses"]
