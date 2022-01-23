FROM golang:1.17.1-alpine3.14 as builder

WORKDIR /api/cmd

COPY . /api

RUN go mod tidy
RUN go build -o balance_api .

FROM alpine:latest

COPY --from=builder api/cmd/balance_api /api/cmd/balance_api

ENTRYPOINT ["api/cmd/balance_api"]