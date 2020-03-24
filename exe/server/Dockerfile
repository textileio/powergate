FROM golang:1.14 as builder

RUN mkdir /app 
WORKDIR /app 

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o powd exe/server/main.go 

FROM alpine
COPY --from=builder /app/api/server/ip2location-ip4.bin /app/ip2location-ip4.bin
COPY --from=builder /app/powd /app/powd
WORKDIR /app 
ENTRYPOINT ["./powd"]