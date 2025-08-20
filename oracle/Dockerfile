FROM rust:1.86.0-alpine AS rust-builder

# Install build dependencies for FROST
RUN apk add --no-cache build-base pkgconfig openssl-dev

# Set working directory
WORKDIR /app
COPY frost/ ./frost/

RUN cargo install cbindgen
# Build FROST library
RUN cd frost/rust && \
    cargo build --release && \
    cbindgen --config cbindgen.toml --crate frost --output ./frost.h

FROM golang:1.23.1-alpine AS go-builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Set working directory
WORKDIR /app

COPY go.work go.work.sum ./
COPY oracle/go.mod oracle/go.sum ./oracle/
COPY lib/go.mod lib/go.sum ./lib/
COPY frost/go.mod frost/go.sum ./frost/

RUN go work edit -dropuse ./relayer
RUN go work edit -dropuse ./indexer
RUN go work edit -dropuse ./metrics

RUN go mod download

COPY frost/ ./frost/
COPY oracle/ ./oracle/
COPY lib/ ./lib/
COPY --from=rust-builder /app/frost/rust/target ./frost/rust/target

# Build frost
RUN cd frost && go build
# Build the Oracle
RUN go build -o ./bin ./oracle/cmd/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies if any
RUN apk add --no-cache ca-certificates libgcc libstdc++

# Copy the binary from builder
COPY --from=go-builder /app/bin /app/

# Set working directory
WORKDIR /app

# Command to run the oracle
CMD ["./bin"] 