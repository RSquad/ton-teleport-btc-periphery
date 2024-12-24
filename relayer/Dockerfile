FROM golang:1.23.1 as builder

WORKDIR /app

COPY go.work go.work.sum ./

RUN go work edit -dropuse ./indexer
RUN go work edit -dropuse ./frost
RUN go work edit -dropuse ./oracle

COPY relayer/go.mod relayer/go.sum ./relayer/
COPY lib/go.mod lib/go.sum ./lib/

RUN go mod download

COPY relayer/ ./relayer/
COPY lib/ ./lib/

RUN go install github.com/go-swagger/go-swagger/cmd/swagger@latest

RUN cd ./lib && ./gen-ton-center-v3-client.sh && cd ..

RUN go build -o ./bin ./relayer/cmd/main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

WORKDIR /app
COPY --from=builder /app/bin /app/

EXPOSE 3000

CMD ["./bin"]
