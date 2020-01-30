FROM golang:1.13.4 as builder
RUN mkdir /app 
WORKDIR /app 
COPY go.mod go.sum ./
COPY extern extern
RUN go mod download
COPY . . 
RUN GOOS=linux go build -o main exe/server/main.go

FROM alpine
RUN apk --no-cache add ca-certificates wget && \
    wget -q -O /etc/apk/keys/sgerrand.rsa.pub https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub && \
    wget https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.28-r0/glibc-2.28-r0.apk && \
    apk add glibc-2.28-r0.apk
COPY --from=builder /app/api/server/ip2location-ip4.bin /app/ip2location-ip4.bin
COPY --from=builder /app/main /app/main
WORKDIR /app 
ENTRYPOINT ["./main"]