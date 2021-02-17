FROM golang:1.16-buster as builder

RUN mkdir /app 
WORKDIR /app 

COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
RUN POW_BUILD_FLAGS="CGO_ENABLED=0 GOOS=linux" make build-powd
RUN POW_BUILD_FLAGS="CGO_ENABLED=0 GOOS=linux" make build-pow

FROM alpine
COPY --from=builder /app/iplocation/maxmind/GeoLite2-City.mmdb /app/GeoLite2-City.mmdb
COPY --from=builder /app/powd /app/powd
COPY --from=builder /app/pow /app/pow
WORKDIR /app 
ENTRYPOINT ["./powd"]
