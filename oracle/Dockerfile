FROM rust:1.86.0-alpine AS rust-builder
RUN apk add --no-cache build-base pkgconfig openssl-dev
WORKDIR /app
COPY frost/ ./frost/
RUN cargo install cbindgen
RUN cd frost/rust && \
    cargo build --release && \
    cbindgen --config cbindgen.toml --crate frost --output ./frost.h

FROM golang:1.23.1-alpine AS go-builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.work go.work.sum ./
COPY oracle/go.mod oracle/go.sum ./oracle/
COPY lib/go.mod lib/go.sum ./lib/
COPY frost/go.mod frost/go.sum ./frost/
RUN go work edit -dropuse ./relayer -dropuse ./indexer
RUN go mod download
COPY frost/ ./frost/
COPY oracle/ ./oracle/
COPY lib/ ./lib/
COPY --from=rust-builder /app/frost/rust/target ./frost/rust/target
RUN cd frost && go build
RUN go build -o ./bin ./oracle/cmd/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates libgcc libstdc++
RUN addgroup -S oracle && adduser -S oracle -G oracle
RUN mkdir -p /app/keystore/secrets /app/keystore/temp && \
    chown -R oracle:oracle /app/keystore && \
    chmod 700 /app/keystore
COPY --from=go-builder /app/bin /app/
WORKDIR /app
USER oracle

CMD ["./bin"] 