FROM golang:latest as builder
WORKDIR /app
COPY main.go /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o main .

FROM scratch
WORKDIR /app
COPY --from=builder /app/main /app
EXPOSE 8080
CMD ["/app/main"]