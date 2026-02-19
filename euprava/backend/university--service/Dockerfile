FROM golang:latest as builder
WORKDIR /app

COPY ./university--service/go.mod ./university--service/go.sum ./
RUN go mod download

COPY ./university--service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8006
CMD ["./main"]
