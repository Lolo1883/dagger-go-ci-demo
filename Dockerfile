# Dockerfile
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN go build -o app .

FROM gcr.io/distroless/static
COPY --from=builder /app/app /
ENTRYPOINT ["/app"]
