# syntax=docker/dockerfile:1

FROM golang:1.23 AS builder

WORKDIR /app

# Copy go.mod/go.sum and download deps
COPY go.mod go.sum ./
RUN go mod download

# Copy all source
COPY . .

# The service to build is passed as a build arg
ARG SERVICE
RUN go build -o /bin/server ./cmd/grpc/${SERVICE}

# Final minimal image
FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /bin/server /app/server

EXPOSE 5005
CMD ["/app/server"]
