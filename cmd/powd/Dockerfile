FROM golang:1.14.2-buster as builder

RUN mkdir /app 
WORKDIR /app 

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o powd cmd/powd/main.go 

FROM alpine
COPY --from=builder /iplocation/maxmind/GeoLite2-City.mmdb /app/GeoLite2-City.mmdb
COPY --from=builder /app/powd /app/powd
WORKDIR /app 
ENTRYPOINT ["./powd"]
