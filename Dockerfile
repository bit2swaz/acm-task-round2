# build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY main.go .
# init mod file locally to keep it clean (or just disable modules for simple scripts)
RUN go env -w GO111MODULE=off
# build a static binary
RUN CGO_ENABLED=0 go build -o proxy main.go

# run stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/proxy .
EXPOSE 80
CMD ["./proxy"]