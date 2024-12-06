FROM golang:1.23.1 as builder

WORKDIR /app

COPY go.work go.work.sum ./

RUN go work edit -dropuse ./relayer
RUN go work edit -dropuse ./frost
COPY indexer/go.mod indexer/go.sum ./indexer/
COPY lib/go.mod lib/go.sum ./lib/

RUN go mod download

COPY indexer/ ./indexer/
COPY lib/ ./lib/

RUN go install github.com/go-swagger/go-swagger/cmd/swagger@latest

RUN cd ./lib && ./gen-ton-center-v3-client.sh && cd ..
RUN cd ./indexer && go generate . && cd ..

RUN go build -o ./bin ./indexer/cmd/main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

WORKDIR /app
COPY --from=builder /app/bin /app/

EXPOSE 3000

CMD ["./bin"]
